package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/geffersonFerraz/frigate-s3-telegram/internal/config"
	"github.com/geffersonFerraz/frigate-s3-telegram/internal/frigate"
	"github.com/geffersonFerraz/frigate-s3-telegram/internal/rabbit"
	"github.com/geffersonFerraz/frigate-s3-telegram/internal/s3"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	redis "github.com/redis/go-redis/v9"
)

const maxSize = 49 * 1024 * 1024 // 50MB

func main() {

	cfg := config.New()

	// Bucket initialization
	s3Client, err := s3.New()
	if err != nil {
		log.Fatalln(err)
	}

	s3Client.CheckAlive()
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()

	s3Bucket, err := s3.Buckets(s3Client.GetClient(), cfg.BUCKET_NAME)
	if err != nil {
		log.Fatalln(err)
	}

	err = s3Bucket.Create(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Bucket created successfully")

	// Prepare startup msg
	startupMsg := "Starting frigate-telegram.\n"
	startupMsg += "Frigate URL:  " + cfg.FrigateURL + "\n"
	log.Println(startupMsg)

	// Telegram initialization
	opts := []bot.Option{}

	b, err := bot.New(cfg.TelegramBotToken, opts...)
	if err != nil {
		log.Fatalln("Error initalizing telegram bot: " + err.Error())
	}
	go b.Start(ctx)

	// Send startup msg. conf.TelegramErrorChatID, startupMsg))
	helloMsg := &bot.SendMessageParams{
		ChatID: cfg.TelegramErrorChatID,
		Text:   startupMsg,
	}
	b.SendMessage(ctx, helloMsg)

	// Frigate initialization
	frigate, err := frigate.NewFrigate()
	if err != nil {
		log.Fatalln(err)
	}

	evts, err := frigate.Events()
	if err != nil {
		log.Fatalln(err)
	}
	for _, x := range evts {
		log.Println(x)
	}

	// RabbitMQ Initialization
	rabbit, err := rabbit.NewRabbitMQ()
	if err != nil {
		log.Fatal(err)
	}
	defer rabbit.Close()

	go func() {
		err = rabbit.Consume(func(msg []byte) error {
			time.Sleep(1 * time.Second)
			event, inProgress, err := frigate.GetEvent(string(msg))

			if err != nil {
				log.Println(err)
				return err
			}
			if inProgress {
				return fmt.Errorf("event %s is still in progress", string(msg))
			}

			var sendBucket = func(filePathClip string) {
				file, err := os.Open(filePathClip)
				if err != nil {
					log.Println(err)
				}
				defer file.Close()

				s3File, err := s3.Files(s3Client.GetClient())
				if err != nil {
					log.Println(err)
				}

				s3File.SetFile(ctx, file)
				s3File.SetBucket(ctx, cfg.BUCKET_NAME)
				timeHumanReadable := time.Unix(int64(event.StartTime), 0).Format("2006-01-02 15:04:05")
				s3File.SetDestinatoin(ctx, event.Camera+"/"+timeHumanReadable+"-"+event.Label+".mp4")
				err = s3File.Upload(ctx)
				if err != nil {
					log.Println(err)
				}

				//
				telegramMessage := &bot.SendMediaGroupParams{
					ChatID:          cfg.TelegramChatID,
					MessageThreadID: getMessageThreadId(event.Camera),
				}
				ctx := context.Background()

				fileName := frigate.SaveThumbnail(*event)
				filePathThumbnail, err := os.Open(fileName)
				if err != nil {
					log.Fatal(err)
				}
				defer filePathThumbnail.Close()
				defer os.Remove(fileName)

				thumb := &models.InputMediaPhoto{
					Media:           "attach://" + fileName,
					MediaAttachment: filePathThumbnail,
					Caption:         "Ended \n " + s3File.GetPresignedURL(ctx),
				}
				medias := []models.InputMedia{
					thumb,
				}

				telegramMessage.Media = medias
				_, err = b.SendMediaGroup(ctx, telegramMessage)
				if err != nil {
					log.Println(err)
				}
			}

			var sendTelegram = func(filePathClip string) {
				file, err := os.Open(filePathClip)
				if err != nil {
					log.Println(err)
				}
				defer file.Close()

				telegramMessage := &bot.SendMediaGroupParams{
					ChatID:          cfg.TelegramChatID,
					MessageThreadID: getMessageThreadId(event.Camera),
				}
				video := &models.InputMediaVideo{
					MediaAttachment: file,
					Media:           "attach://" + filePathClip,
					Caption:         event.Camera + " Event: " + event.Label + ", ID: " + event.ID,
				}

				medias := []models.InputMedia{
					video,
				}

				telegramMessage.Media = medias
				_, err = b.SendMediaGroup(ctx, telegramMessage)
				if err != nil {
					log.Println(err)
				}
			}

			go func() {
				log.Printf("Received: %s\n", string(msg))
				time.Sleep(60 * time.Second)

				if event.HasClip {
					filePathClip := frigate.SaveClip(*event)

					file, err := os.Open(filePathClip)
					if err != nil {
						log.Println(err)
					}

					fileInfo, err := file.Stat()
					if err != nil {
						log.Println(err)
					}

					waitGroup := sync.WaitGroup{}

					if fileInfo.Size() > maxSize {
						waitGroup.Add(1)
						go func() {
							defer waitGroup.Done()
							sendBucket(filePathClip)
						}()

					} else {
						// waitGroup.Add(1)
						// go func() {
						// 	defer waitGroup.Done()
						// 	sendBucket(filePathClip)
						// }()

						waitGroup.Add(1)
						go func() {
							defer waitGroup.Done()
							sendTelegram(filePathClip)
						}()
					}

					waitGroup.Wait()

					defer file.Close()
					defer os.Remove(filePathClip)

				}
			}()

			return nil
		})
	}()

	// Redis
	var rdb = redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword, // no password set
		DB:       cfg.RedisDB,       // use default DB
		Protocol: cfg.RedisProtocol, // specify 2 for RESP 2 or 3 for RESP 3
	})

	// Create a go routine channel to send a bot message, when evts is not empty
	for {
		evts, err = frigate.Events()
		time.Sleep(200 * time.Millisecond)
		if err != nil {
			log.Fatalln(err)
		}
		if len(evts) > 0 {
			go func() {
				for _, x := range evts {
					if rdb.Get(ctx, x.ID).Val() == x.ID {
						continue
					}
					err = rdb.Set(ctx, x.ID, x.ID, 24*time.Hour).Err()

					rabbit.Publish(ctx, []byte(x.ID))

					//

					telegramMessage := &bot.SendMediaGroupParams{
						ChatID:          cfg.TelegramChatID,
						MessageThreadID: getMessageThreadId(x.Camera),
					}
					ctx := context.Background()

					//Convert image base64 from x.Thumbnail to image and attach to telegram message
					fileName := frigate.SaveThumbnail(x)
					filePathThumbnail, err := os.Open(fileName)
					if err != nil {
						log.Fatal(err)
					}
					defer filePathThumbnail.Close()
					defer os.Remove(fileName)

					thumb := &models.InputMediaPhoto{
						Media:           "attach://" + fileName,
						MediaAttachment: filePathThumbnail,
						Caption:         x.Camera + " Event: " + x.Label + ", ID: " + x.ID,
					}
					medias := []models.InputMedia{
						thumb,
					}

					telegramMessage.Media = medias
					_, err = b.SendMediaGroup(ctx, telegramMessage)
					if err != nil {
						log.Println(err)
					}
				}
			}()
		}
	}
}

func getMessageThreadId(camera string) int {
	threadList := make(map[string]int)
	threadList["General"] = 0
	threadList["Bolacha"] = 2
	threadList["Rua"] = 3
	threadList["Tras"] = 4
	threadList["RuaMAto"] = 5
	threadList["Portao"] = 26
	threadList["TrasPorta"] = 366
	return threadList[camera]
}
