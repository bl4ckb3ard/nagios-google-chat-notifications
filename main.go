package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"log"

	"google.golang.org/api/chat/v1"
)

func main() {
	alertType := flag.String("type", "", "Type of alert: 'host' or 'service'")
	webhookURL := flag.String("webhook", "", "Google Chat Webhook URL")
	nagiosURL := flag.String("nagios-url", "", "Base URL for Nagios web interface")

	notificationType := flag.String("notification-type", "", "Nagios $NOTIFICATIONTYPE$")
	hostName := flag.String("hostname", "", "Nagios $HOSTNAME$")
	hostAddress := flag.String("hostaddress", "", "Nagios $HOSTADDRESS$")
	state := flag.String("state", "", "Nagios $HOSTSTATE$ or $SERVICESTATE$")
	serviceDesc := flag.String("service-desc", "", "Nagios $SERVICEDESC$")
	output := flag.String("output", "", "Nagios $HOSTOUTPUT$ or $SERVICEOUTPUT$")
	dateTime := flag.String("datetime", "", "Nagios $LONGDATETIME$")

	author := flag.String("author", "", "Nagios $NOTIFICATIONAUTHOR$")
	comment := flag.String("comment", "", "Nagios $NOTIFICATIONCOMMENT$")

	mentionAll := flag.Bool("mention-all", false, "Include <users/all> to ping everyone in the space")

	flag.Parse()

	if *webhookURL == "" || *alertType == "" {
		log.Fatal("Error: --webhook and --type are required.")
	}

	emoji := "💬"
	upperType := strings.ToUpper(*notificationType)
	upperState := strings.ToUpper(*state)

	if upperType == "ACKNOWLEDGEMENT" {
		emoji = "✅"
	} else if strings.Contains(upperType, "DOWNTIME") {
		emoji = "💤"
	} else if strings.Contains(upperType, "FLAPPING") {
		emoji = "🔀"
	} else if upperType == "RECOVERY" || upperState == "OK" || upperState == "UP" {
		emoji = "🟢"
	} else if upperState == "WARNING" {
		emoji = "⚠️"
	} else if upperState == "CRITICAL" || upperState == "DOWN" || upperType == "PROBLEM" {
		emoji = "🚨"
	} else if upperState == "UNKNOWN" || upperState == "UNREACHABLE" {
		emoji = "❓"
	}

	var title, subtitle, link, ackLink string
	escapedHost := url.QueryEscape(*hostName)
	var widgets []*chat.GoogleAppsCardV1Widget

	if strings.ToLower(*alertType) == "service" {
		title = fmt.Sprintf("%s %s: %s", emoji, *notificationType, *serviceDesc)
		subtitle = fmt.Sprintf("Alert triggered on: %s", *hostName)
		escapedService := url.QueryEscape(*serviceDesc)

		link = fmt.Sprintf("%s/cgi-bin/extinfo.cgi?type=2&host=%s&service=%s", *nagiosURL, escapedHost, escapedService)
		ackLink = fmt.Sprintf("%s/cgi-bin/cmd.cgi?cmd_typ=34&host=%s&service=%s", *nagiosURL, escapedHost, escapedService)

		widgets = append(widgets,
			&chat.GoogleAppsCardV1Widget{DecoratedText: &chat.GoogleAppsCardV1DecoratedText{
				TopLabel: "Target Host",
				Text:     fmt.Sprintf("<b><font color=\"#1a73e8\">%s</font></b>", *hostName),
			}},
			&chat.GoogleAppsCardV1Widget{DecoratedText: &chat.GoogleAppsCardV1DecoratedText{
				TopLabel: "State",
				Text:     fmt.Sprintf("<b>%s</b>", *state),
			}},
			&chat.GoogleAppsCardV1Widget{DecoratedText: &chat.GoogleAppsCardV1DecoratedText{
				TopLabel: "Host Address",
				Text:     *hostAddress,
			}},
			&chat.GoogleAppsCardV1Widget{DecoratedText: &chat.GoogleAppsCardV1DecoratedText{
				TopLabel: "Additional Info",
				Text:     *output,
			}},
		)

	} else if strings.ToLower(*alertType) == "host" {
		title = fmt.Sprintf("%s %s: Host %s", emoji, *notificationType, *hostName)
		subtitle = fmt.Sprintf("Address: %s", *hostAddress)

		link = fmt.Sprintf("%s/cgi-bin/extinfo.cgi?type=1&host=%s", *nagiosURL, escapedHost)
		ackLink = fmt.Sprintf("%s/cgi-bin/cmd.cgi?cmd_typ=33&host=%s", *nagiosURL, escapedHost)

		widgets = append(widgets,
			&chat.GoogleAppsCardV1Widget{DecoratedText: &chat.GoogleAppsCardV1DecoratedText{
				TopLabel: "Target Host",
				Text:     fmt.Sprintf("<b><font color=\"#1a73e8\">%s</font></b>", *hostName),
			}},
			&chat.GoogleAppsCardV1Widget{DecoratedText: &chat.GoogleAppsCardV1DecoratedText{
				TopLabel: "State",
				Text:     fmt.Sprintf("<b>%s</b>", *state),
			}},
			&chat.GoogleAppsCardV1Widget{DecoratedText: &chat.GoogleAppsCardV1DecoratedText{
				TopLabel: "Additional Info",
				Text:     *output,
			}},
		)
	} else {
		log.Fatal("Error: --type must be either 'host' or 'service'")
	}

	if (upperType == "ACKNOWLEDGEMENT" || strings.Contains(upperType, "DOWNTIME")) && *author != "" {
		
		authorLabel := "Acknowledged By"
		if strings.Contains(upperType, "DOWNTIME") {
			authorLabel = "Scheduled By"
		}

		widgets = append(widgets,
			&chat.GoogleAppsCardV1Widget{DecoratedText: &chat.GoogleAppsCardV1DecoratedText{TopLabel: authorLabel, Text: *author}},
			&chat.GoogleAppsCardV1Widget{DecoratedText: &chat.GoogleAppsCardV1DecoratedText{TopLabel: "Comment", Text: *comment}},
		)
	}

	widgets = append(widgets, &chat.GoogleAppsCardV1Widget{DecoratedText: &chat.GoogleAppsCardV1DecoratedText{TopLabel: "Time", Text: *dateTime}})

	buttons := []*chat.GoogleAppsCardV1Button{
		{
			Text: "View Details",
			OnClick: &chat.GoogleAppsCardV1OnClick{
				OpenLink: &chat.GoogleAppsCardV1OpenLink{Url: link},
			},
		},
	}

	if upperType == "PROBLEM" {
		buttons = append(buttons, &chat.GoogleAppsCardV1Button{
			Text: "Acknowledge",
			OnClick: &chat.GoogleAppsCardV1OnClick{
				OpenLink: &chat.GoogleAppsCardV1OpenLink{Url: ackLink},
			},
		})
	}

	widgets = append(widgets, &chat.GoogleAppsCardV1Widget{
		ButtonList: &chat.GoogleAppsCardV1ButtonList{
			Buttons: buttons,
		},
	})

	payload := &chat.Message{
		CardsV2: []*chat.CardWithId{
			{
				CardId: "nagios-alert",
				Card: &chat.GoogleAppsCardV1Card{
					Header: &chat.GoogleAppsCardV1CardHeader{
						Title:    title,
						Subtitle: subtitle,
					},
					Sections: []*chat.GoogleAppsCardV1Section{
						{Widgets: widgets},
					},
				},
			},
		},
	}

	if *mentionAll {
		payload.Text = "<users/all>"
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v\n", err)
	}

	resp, err := http.Post(*webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error sending request: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Fatalf("Failed to send message. HTTP Status: %d\n", resp.StatusCode)
	}

	log.Println("Card notification sent successfully.")
}
