package message

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/qj0r9j0vc2/kko/internal/kakao"
)

const (
	sendToMeURL     = "https://kapi.kakao.com/v2/api/talk/memo/default/send"
	sendToFriendURL = "https://kapi.kakao.com/v1/api/talk/friends/message/default/send"
)

type Service struct {
	client *kakao.Client
}

func NewService(client *kakao.Client) *Service {
	return &Service{client: client}
}

func (s *Service) SendToMe(ctx context.Context, template *TemplateObject) error {
	templateJSON, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("marshal template: %w", err)
	}

	form := url.Values{}
	form.Set("template_object", string(templateJSON))

	_, err = s.client.Post(ctx, sendToMeURL, form, kakao.WithOAuth())
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	return nil
}

func (s *Service) SendToFriend(ctx context.Context, receiverUUID string, template *TemplateObject) error {
	templateJSON, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("marshal template: %w", err)
	}

	form := url.Values{}
	form.Set("receiver_uuids", fmt.Sprintf(`["%s"]`, receiverUUID))
	form.Set("template_object", string(templateJSON))

	_, err = s.client.Post(ctx, sendToFriendURL, form, kakao.WithOAuth())
	if err != nil {
		return fmt.Errorf("send message to friend: %w", err)
	}
	return nil
}
