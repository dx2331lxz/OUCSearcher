package tools

import (
	"OUCSearcher/models"
	"log"
)

func UpdateCrawDone() error {
	err := models.DeleteAllVisitedUrls()
	if err != nil {
		log.Println("Error deleting all visited urls:", err)
		return err
	}
	err = models.SetCrawDoneToZero()
	if err != nil {
		log.Println("Error setting craw_done to zero:", err)
		return err
	}
	return nil
}
