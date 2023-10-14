package repositories

import (
	"github.com/gin-gonic/gin"
)

type TransactionData struct {
	Src                int64   `form:"src"`
	Prc                int64   `form:"prc"`
	Ord                string  `form:"Ord"`
	Holder             string  `form:"Holder"`
	SuccessCode        int64   `form:"successcode"`
	Ref                string  `form:"Ref"`
	PayRef             int64   `form:"PayRef"`
	Amt                float64 `form:"Amt"`
	Cur                string  `form:"Cur"`
	Remark             string  `form:"remark"`
	AuthId             string  `form:"AuthId"`
	Eci                string  `form:"Eci"`
	PayerAuth          string  `form:"payerAuth"`
	SourceIp           string  `form:"sourceIp"`
	IpCountry          string  `form:"ipCountry"`
	PayMethod          string  `form:"payMethod"`
	CardIssuingCountry string  `form:"cardIssuingCountry"`
	ChannelType        string  `form:"channelType"`
	SecureHash         string  `form:"secureHash"`
	AlertCode          string  `form:"AlertCode"`
	MerchantId         int64   `form:"MerchantId"`
	TxTime             string  `form:"TxTime"`
}

func (t *TransactionData) SetFromReq(c *gin.Context) error {
	return c.ShouldBind(t)
}
