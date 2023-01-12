// Copyright Epic Games, Inc. All Rights Reserved.

package handler

import (
	"fmt"
	"github.com/ashwanthkumar/slack-go-webhook"
	"net/http"
)

func (s *SSRFSheriffRouter) PostNotification(r *http.Request) {
	payload := slack.Payload{
		Username: "SSRF Sheriff",
		Text:     fmt.Sprintf("SSRF Hit from IP `%s` on path `%s` with headers `%s`", r.RemoteAddr, r.URL.Path, r.Header),
		Markdown: true,
	}

	err := slack.Send(s.webhook, "", payload)
	if err != nil {
		s.logger.Error(fmt.Sprintf("failed to send slack message: %v", err))
	}

}
