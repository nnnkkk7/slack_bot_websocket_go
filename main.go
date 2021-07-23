package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	"github.com/slack-go/slack"
)

func main() {
	webApi := slack.New(
		os.Getenv("SLACK_BOT_TOKEN"),
		slack.OptionAppLevelToken(os.Getenv("SLACK_APP_TOKEN")),
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
	)
	socketMode := socketmode.New(
		webApi,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "sm: ", log.Lshortfile|log.LstdFlags)),
	)
	authTest, authTestErr := webApi.AuthTest()
	if authTestErr != nil {
		fmt.Fprintf(os.Stderr, "SLACK_BOT_TOKEN is invalid: %v\n", authTestErr)
		os.Exit(1)
	}
	selfUserId := authTest.UserID
	fmt.Println(selfUserId)

	go func() {
		for envelope := range socketMode.Events {
			switch envelope.Type {
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, err := slackevents.ParseEvent(envelope.Request.Payload, slackevents.OptionNoVerifyToken())
				if err != nil {
					fmt.Println("Something went wrong while parsing envelope")
					return
				}
				innerEvent := eventsAPIEvent.InnerEvent
				//var res *slackevents.ChallengeResponse
				switch ev := innerEvent.Data.(type) {
				case *slackevents.AppMentionEvent:
					fmt.Println("----------Yay! It worked!----------")
					if strings.Contains(ev.Text, "こんにちは") {
						_, _, err := webApi.PostMessage(
							ev.Channel,
							slack.MsgOptionText(
								fmt.Sprintf(":wave: こんにちは <@%v> さん！", ev.User),
								false,
							),
						)
						if err != nil {
							log.Printf("Failed to reply: %v", err)
						}
					}

				}
			default:
				socketMode.Debugf("Skipped: %v", envelope.Type)
			}
		}
	}()

	socketMode.Run()
}
