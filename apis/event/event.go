package apis

import (
	"naotodoserver/models"
	"strconv"
)

func ToEventResponseList(events []models.Event) []models.EventResponse {
	responseList := make([]models.EventResponse, len(events))
	for i, item := range events {
		responseList[i] = ToEventResponse(&item)
	}
	return responseList
}

func ToEventResponse(event *models.Event) models.EventResponse {
	var eventResponse models.EventResponse
	eventResponse.ID = strconv.FormatInt(event.ID, 10)
	eventResponse.CreatedAt = event.CreatedAt
	eventResponse.UpdatedAt = event.UpdatedAt
	eventResponse.DeletedAt = event.DeletedAt

	eventResponse.UserId = strconv.FormatInt(event.UserId, 10)
	eventResponse.TodoId = strconv.FormatInt(event.TodoId, 10)

	eventResponse.Name = event.Name
	eventResponse.Description = event.Description
	eventResponse.IsDone = event.IsDone
	eventResponse.SortId = event.SortId

	return eventResponse
}
