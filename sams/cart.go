package sams

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
)

type FloorInfo struct {
	FloorId         int64         `json:"floorId"`
	Amout           string        `json:"amout"`
	Quantity        int64         `json:"quantity"`
	StoreInfo       StoreInfo     `json:"storeInfo"`
	NormalGoodsList []NormalGoods `json:"normalGoodsList"`
}

type Cart struct {
	DeliveryAddress Address     `json:"deliveryAddress"`
	FloorInfoList   []FloorInfo `json:"floorInfoList"`
}

func (session *Session) parseFloorInfo(result gjson.Result) (error, FloorInfo) {
	floorInfo := FloorInfo{}
	floorInfo.FloorId = result.Get("floorId").Int()
	floorInfo.Amout = result.Get("amount").Str
	floorInfo.Quantity = result.Get("quantity").Int()
	floorInfo.StoreInfo = StoreInfo{
		StoreId:                 result.Get("storeInfo.storeId").Str,
		StoreType:               result.Get("storeInfo.storeType").Int(),
		AreaBlockId:             result.Get("storeInfo.areaBlockId").Str,
		StoreDeliveryTemplateId: result.Get("storeInfo.storeDeliveryTemplateId").Str,
		DeliveryModeId:          result.Get("storeInfo.deliveryModeId").Str,
	}

	for _, v := range result.Get("normalGoodsList").Array() {
		_, normalGoods := session.parseNormalGoods(v)
		floorInfo.NormalGoodsList = append(floorInfo.NormalGoodsList, normalGoods)
	}

	return nil, floorInfo
}

func (session *Session) parseMiniProgramGoodsInfo(result gjson.Result) (error, FloorInfo) {
	floorInfo := FloorInfo{}
	floorInfo.FloorId = session.FloorId
	floorInfo.Amout = result.Get("selectedAmount").Str
	for _, v := range result.Get("normalGoodsList").Array() {
		_, normalGoods := session.parseNormalGoods(v)
		floorInfo.NormalGoodsList = append(floorInfo.NormalGoodsList, normalGoods)
		for _, s := range session.StoreList {
			if normalGoods.StoreId == s.StoreId {
				floorInfo.StoreInfo = StoreInfo{
					StoreId:                 s.StoreId,
					StoreType:               s.StoreType,
					AreaBlockId:             s.AreaBlockId,
					StoreDeliveryTemplateId: s.StoreDeliveryTemplateId,
					DeliveryModeId:          s.DeliveryModeId,
				}
			}
		}
	}
	return nil, floorInfo
}

func (session *Session) SetCartInfo(result gjson.Result) error {
	cart := Cart{}
	cart.FloorInfoList = make([]FloorInfo, 0)
	switch session.Setting.DeviceType {
	case 1:
		for _, v := range result.Get("data.floorInfoList").Array() {
			_, floor := session.parseFloorInfo(v)
			cart.FloorInfoList = append(cart.FloorInfoList, floor)
		}
	case 2:
		for _, v := range result.Get("data.miniProgramGoodsInfo").Array() {
			_, floor := session.parseMiniProgramGoodsInfo(v)
			cart.FloorInfoList = append(cart.FloorInfoList, floor)
		}
	default:
		return errors.New("未知设备类型")
	}
	session.Cart = cart
	return nil
}

func (session *Session) CheckCart() error {
	session.Cart = Cart{}
	data := CartParam{
		StoreList:         session.StoreList,
		AddressId:         session.Address.AddressId,
		Uid:               "",
		DeliveryType:      fmt.Sprintf("%d", session.Setting.DeliveryType),
		HomePagelatitude:  session.Address.Latitude,
		HomePagelongitude: session.Address.Longitude,
	}
	dataStr, _ := json.Marshal(data)
	err, result := session.Request.POST(CartAPI, dataStr)
	if err != nil {
		return err
	}
	return session.SetCartInfo(result)
}