package entities

import "time"

type Session struct {
	Id         int64     `json:"id"`
	UserId     int64     `json:"userId"`
	Token      string    `json:"token"`
	ExpiredAt  time.Time `json:"expiredAt"`
	DeviceType string    `json:"deviceType"`
}

func (s *Session) IsIdValid() bool {
	return s != nil && s.Id > 0
}

func (s *Session) IsExpired() bool {
	return s != nil && s.ExpiredAt.Before(time.Now())
}

func (s *Session) IsValid() bool {
	return s.IsIdValid() && !s.IsExpired()
}
