package page

import (
	"github.com/google/uuid"
	"mindlink.io/mindlink/pkg/models"
)

type respUserPage struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	NodeNum       int       `json:"nodeNum"`
	ConnectionNum int       `json:"connectionNum"`
}

type respListUserPages []respUserPage

type respModifyUserPage struct {
	SuccessOK bool `json:"success_ok"`
}

func convertPageIntoResp(page *models.Page) respUserPage {
	return respUserPage{
		ID:            page.ID,
		Name:          page.Name,
		NodeNum:       page.NodeNum,
		ConnectionNum: page.ConnectionNum,
	}
}
