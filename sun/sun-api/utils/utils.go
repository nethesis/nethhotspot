/*
 * Copyright (C) 2017 Nethesis S.r.l.
 * http://www.nethesis.it - info@nethesis.it
 *
 * This file is part of Icaro project.
 *
 * Icaro is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License,
 * or any later version.
 *
 * Icaro is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Icaro.  If not, see COPYING.
 *
 * author: Edoardo Spadoni <edoardo.spadoni@nethesis.it>
 */

package utils

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/nethesis/icaro/sun/sun-api/configuration"
	"github.com/nethesis/icaro/sun/sun-api/database"
	"github.com/nethesis/icaro/sun/sun-api/defaults"
	"github.com/nethesis/icaro/sun/sun-api/models"
)

func SetDefaultHotspotPreferences(hotspotId int) {
	db := database.Instance()

	// iterate all default hotspot preferences
	for k, v := range defaults.HotspotPreferences {
		hsPreferences := models.HotspotPreference{
			HotspotId: hotspotId,
			Key:       k,
			Value:     v,
		}
		db.Save(&hsPreferences)
	}

	// set captive portal defaults from configuration
	db.Save(&models.HotspotPreference{HotspotId: hotspotId, Key: "captive_1_redir", Value: configuration.Config.CaptivePortal.Redirect})
	db.Save(&models.HotspotPreference{HotspotId: hotspotId, Key: "captive_2_title", Value: configuration.Config.CaptivePortal.Title})
	db.Save(&models.HotspotPreference{HotspotId: hotspotId, Key: "captive_3_logo", Value: configuration.Config.CaptivePortal.LogoContents})
	db.Save(&models.HotspotPreference{HotspotId: hotspotId, Key: "captive_4_subtitle", Value: configuration.Config.CaptivePortal.Subtitle})
	db.Save(&models.HotspotPreference{HotspotId: hotspotId, Key: "captive_5_banner", Value: configuration.Config.CaptivePortal.BannerContents})
	db.Save(&models.HotspotPreference{HotspotId: hotspotId, Key: "captive_6_description", Value: configuration.Config.CaptivePortal.Description})
	db.Save(&models.HotspotPreference{HotspotId: hotspotId, Key: "captive_7_background", Value: configuration.Config.CaptivePortal.Background})

}

func OffsetCalc(page string, limit string) [2]int {
	var resLimit = 0
	var resOffset = 0

	if len(page) == 0 {
		page = "1"
	}
	if len(limit) == 0 {
		limit = "-1"
	}

	limitInt, errLimit := strconv.Atoi(limit)
	if errLimit != nil {
		fmt.Println(errLimit.Error())
	}

	pageInt, errPage := strconv.Atoi(page)
	if errPage != nil {
		fmt.Println(errPage.Error())
	}

	resLimit = limitInt
	resOffset = (pageInt - 1) * resLimit

	result := [2]int{resOffset, resLimit}
	return result
}

func ExtractToken(token string) models.AccessToken {
	var accessToken models.AccessToken
	db := database.Instance()
	db.Where("token = ?", token).First(&accessToken)

	return accessToken
}

func DeleteToken(token string) {
	var accessToken models.AccessToken
	db := database.Instance()
	db.Where("token = ?", token).First(&accessToken)

	db.Delete(&accessToken)
}

func RefreshToken(token string) {
	var accessToken models.AccessToken
	db := database.Instance()
	db.Where("token = ?", token).First(&accessToken)

	// add 1 day to expiration date
	accessToken.Expires = time.Now().UTC().AddDate(0, 0, configuration.Config.TokenExpiresDays)
	db.Save(&accessToken)

}

func ExtractHotspotIds(accountId int, admin bool, hotspotId int) []int {
	var hotspots []models.Hotspot
	account := GetAccountById(accountId)

	db := database.Instance()

	if account.Type == "customer" || account.Type == "desk" {
		accountId = account.CreatorId

		// retrieve hotspot of this account
		var accountsHotspot models.AccountsHotspot
		db.Where("account_id = ?", account.Id).First(&accountsHotspot)
		hotspotId = accountsHotspot.HotspotId
	}

	if admin {
		if hotspotId != 0 {
			db.Select("id").Where("id = ?", hotspotId).Find(&hotspots)
		} else {
			db.Select("id").Find(&hotspots)
		}
	} else {
		if hotspotId != 0 {
			db.Select("id").Where("account_id = ? AND id = ?", accountId, hotspotId).Find(&hotspots)
		} else {
			if account.Type == "reseller" {
				db.Select("id").Where("account_id = ?", accountId).Find(&hotspots)
			}
		}
	}

	result := []int{}

	for _, hotspot := range hotspots {
		result = append(result, hotspot.Id)
	}

	return result
}

func GetAccountById(id int) models.Account {
	var account models.Account
	db := database.Instance()
	db.Where("id = ?", id).First(&account)

	return account
}

func GetHotspotById(id string) models.Hotspot {
	var hotspot models.Hotspot
	db := database.Instance()
	db.Where("id = ?", id).First(&hotspot)

	return hotspot
}

func GetHotspotByName(name string) models.Hotspot {
	var hotspot models.Hotspot
	db := database.Instance()
	db.Where("name = ?", name).First(&hotspot)

	return hotspot
}

func HotspotIsOverQuota(hotspotId int) bool {
	var hotspot models.Hotspot
	var subscription models.Subscription
	var count int
	db := database.Instance()
	db.Set("gorm:auto_preload", true)
	db.Preload("Account").Where("id = ?", hotspotId).First(&hotspot)
	db.Preload("SubscriptionPlan").Where("account_id = ?", hotspot.Account.Id).First(&subscription)

	query := fmt.Sprintf("SELECT COUNT(units.id) as count FROM units JOIN hotspots on units.hotspot_id = hotspots.id WHERE hotspots.account_id = %d", hotspot.Account.Id)
	db.Raw(query).Count(&count)

	return count >= subscription.SubscriptionPlan.MaxUnits
}

func CanChangeCaptivePortalOptions(accountId int) bool {
	var subscription models.Subscription

	// get account
	account := GetAccountById(accountId)

	db := database.Instance()
	db.Set("gorm:auto_preload", true)

	if account.Type == "customer" || account.Type == "desk" {
		db.Preload("SubscriptionPlan").Where("account_id = ?", account.CreatorId).First(&subscription)
	} else {
		db.Preload("SubscriptionPlan").Where("account_id = ?", accountId).First(&subscription)

	}

	return subscription.SubscriptionPlan.WingsCustomization
}

func GetAccountOrCreate(claims map[string]interface{}) models.Account {
	var account models.Account
	var subscriptionPlan models.SubscriptionPlan
	var accountSMS models.AccountSmsCount

	//get account
	db := database.Instance()
	db.Where("username = ?", claims["sub"].(string)).First(&account)

	if account.Id == 0 {
		newAccount := models.Account{
			CreatorId: 1, //admin
			Uuid:      claims["sub"].(string),
			Type:      "reseller",
			Name:      claims["sub"].(string),
			Username:  claims["sub"].(string),
			Password:  "",
			Email:     claims["sub"].(string),
			Created:   time.Now().UTC(),
		}
		db.Save(&newAccount)

		// retrieve subscription plan free
		db.Where("id = 1").First(&subscriptionPlan)

		// create new subscription
		subscription := models.Subscription{
			AccountId:          newAccount.Id,
			SubscriptionPlanId: 1, // free
			ValidFrom:          time.Now().UTC(),
			ValidUntil:         time.Now().UTC().AddDate(0, 0, subscriptionPlan.Period),
			Created:            time.Now().UTC(),
		}
		db.Save(&subscription)

		// create SMS accounting
		accountSMS.AccountId = newAccount.Id
		accountSMS.SmsMaxCount = subscriptionPlan.IncludedSMS
		db.Save(&accountSMS)

		return newAccount
	}
	return account
}

func GetSubscriptionPlanByCode(code string) models.SubscriptionPlan {
	var subscriptionPlan models.SubscriptionPlan
	db := database.Instance()
	db.Where("code = ?", code).First(&subscriptionPlan)

	return subscriptionPlan
}

func GetSubscriptionPlanById(id int) models.SubscriptionPlan {
	var subscriptionPlan models.SubscriptionPlan
	db := database.Instance()
	db.Where("id = ?", id).First(&subscriptionPlan)

	return subscriptionPlan
}

func Round(val float64, roundOn float64, places int) float64 {
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)

	var round float64
	if val > 0 {
		if div >= roundOn {
			round = math.Ceil(digit)
		} else {
			round = math.Floor(digit)
		}
	} else {
		if div >= roundOn {
			round = math.Floor(digit)
		} else {
			round = math.Ceil(digit)
		}
	}
	return round / pow
}

func Contains(intSlice []int, searchInt int) bool {
	for _, value := range intSlice {
		if value == searchInt {
			return true
		}
	}
	return false
}
