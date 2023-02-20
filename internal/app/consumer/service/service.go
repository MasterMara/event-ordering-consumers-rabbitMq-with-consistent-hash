package service

import "event-ordering-consumers-rabbitmq-with-consistent-hash/internal/app/consumer/repository"

type Service interface {
}

type service struct {
	repository repository.Repository
}

func NewService(repository repository.Repository) Service {
	return &service{
		repository: repository,
	}
}
