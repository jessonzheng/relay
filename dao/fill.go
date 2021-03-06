/*

  Copyright 2017 Loopring Project Ltd (Loopring Foundation).

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.

*/

package dao

import (
	"fmt"
	"github.com/Loopring/relay/types"
	"github.com/ethereum/go-ethereum/common"
)

type FillEvent struct {
	ID            int    `gorm:"column:id;primary_key;" json:"id"`
	Protocol      string `gorm:"column:contract_address;type:varchar(42)" json:"protocol"`
	Owner         string `gorm:"column:owner;type:varchar(42)" json:"owner"`
	RingIndex     int64  `gorm:"column:ring_index;" json:"ringIndex"`
	BlockNumber   int64  `gorm:"column:block_number" json:"blockNumber"`
	CreateTime    int64  `gorm:"column:create_time" json:"createTime"`
	RingHash      string `gorm:"column:ring_hash;varchar(82)" json:"ringHash"`
	FillIndex     int64  `gorm:"column:fill_index" json:"fillIndex"`
	TxHash        string `gorm:"column:tx_hash;type:varchar(82)" json:"txHash"`
	PreOrderHash  string `gorm:"column:pre_order_hash;varchar(82)" json:"preOrderHash"`
	NextOrderHash string `gorm:"column:next_order_hash;varchar(82)" json:"nextOrderHash"`
	OrderHash     string `gorm:"column:order_hash;type:varchar(82)" json:"orderHash"`
	AmountS       string `gorm:"column:amount_s;type:varchar(30)" json:"amountS"`
	AmountB       string `gorm:"column:amount_b;type:varchar(30)" json:"amountB"`
	TokenS        string `gorm:"column:token_s;type:varchar(42)" json:"tokenS"`
	TokenB        string `gorm:"column:token_b;type:varchar(42)" json:"tokenB"`
	LrcReward     string `gorm:"column:lrc_reward;type:varchar(30)" json:"lrcReward"`
	LrcFee        string `gorm:"column:lrc_fee;type:varchar(30)" json:"lrcFee"`
	SplitS        string `gorm:"column:split_s;type:varchar(30)" json:"splitS"`
	SplitB        string `gorm:"column:split_b;type:varchar(30)" json:"splitB"`
	Market        string `gorm:"column:market;type:varchar(42)" json:"market"`
}

// convert chainclient/orderFilledEvent to dao/fill
func (f *FillEvent) ConvertDown(src *types.OrderFilledEvent) error {
	f.AmountS = src.AmountS.String()
	f.AmountB = src.AmountB.String()
	f.LrcReward = src.LrcReward.String()
	f.LrcFee = src.LrcFee.String()
	f.SplitS = src.SplitS.String()
	f.SplitB = src.SplitB.String()
	f.Protocol = src.ContractAddress.Hex()
	f.RingIndex = src.RingIndex.Int64()
	f.BlockNumber = src.Blocknumber.Int64()
	f.CreateTime = src.Time.Int64()
	f.RingHash = src.Ringhash.Hex()
	f.TxHash = src.TxHash.Hex()
	f.PreOrderHash = src.PreOrderHash.Hex()
	f.NextOrderHash = src.NextOrderHash.Hex()
	f.OrderHash = src.OrderHash.Hex()
	f.TokenS = src.TokenS.Hex()
	f.TokenB = src.TokenB.Hex()
	f.Owner = src.Owner.Hex()
	f.FillIndex = src.FillIndex.Int64()
	f.Market = src.Market

	return nil
}

func (s *RdsServiceImpl) FindFillEventByRinghashAndOrderhash(ringhash, orderhash common.Hash) (*FillEvent, error) {
	var (
		fill FillEvent
		err  error
	)
	err = s.db.Where("ring_hash = ? and order_hash = ?", ringhash.Hex(), orderhash.Hex()).First(&fill).Error

	return &fill, err
}

func (s *RdsServiceImpl) FillsPageQuery(query map[string]interface{}, pageIndex, pageSize int) (res PageResult, err error) {
	fills := make([]FillEvent, 0)
	res = PageResult{PageIndex: pageIndex, PageSize: pageSize, Data: make([]interface{}, 0)}
	err = s.db.Where(query).Order("create_time desc").Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&fills).Error
	if err != nil {
		return res, err
	}
	err = s.db.Model(&FillEvent{}).Where(query).Count(&res.Total).Error
	if err != nil {
		return res, err
	}

	for _, fill := range fills {
		res.Data = append(res.Data, fill)
	}
	return
}

func (s *RdsServiceImpl) QueryRecentFills(market, owner string, start int64, end int64) (fills []FillEvent, err error) {

	query := make(map[string]interface{})

	if market != "" {
		query["market"] = market
	}

	if owner != "" {
		query["owner"] = owner
	}

	timeQuery := buildTimeQueryString(start, end)

	if timeQuery != "" {
		err = s.db.Where(query).Where(timeQuery).Order("create_time desc").Limit(100).Find(&fills).Error
	} else {
		err = s.db.Where(query).Order("create_time desc").Limit(100).Find(&fills).Error
	}
	return
}

func buildTimeQueryString(start, end int64) string {
	rst := ""
	if start != 0 && end == 0 {
		rst += "create_time >= " + fmt.Sprintf("%v", start)
	} else if start != 0 && end != 0 {
		rst += "create_time >= " + fmt.Sprintf("%v", start) + " AND create_time <= " + fmt.Sprintf("%v", end)
	} else if start == 0 && end != 0 {
		rst += "create_time <= " + fmt.Sprintf("%v", end)
	}
	return rst
}

func (s *RdsServiceImpl) RollBackFill(from, to int64) error {
	return s.db.Where("block_number > ? and block_number <= ?", from, to).Delete(&FillEvent{}).Error
}
