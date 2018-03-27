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

package methods

import (
	"net/http"
	"strconv"
	"time"

	"github.com/nethesis/icaro/wax/utils"

	"github.com/gin-gonic/gin"

	"github.com/nethesis/icaro/sun/sun-api/database"
	"github.com/nethesis/icaro/sun/sun-api/methods"
	"github.com/nethesis/icaro/sun/sun-api/models"
)

type smsMarketingData struct {
	Number string `json:"number"`
}
type emailMarketingData struct {
	Email string `json:"email"`
}

func SMSAuth(c *gin.Context) {
	number := c.Param("number")
	digest := c.Query("digest")
	uuid := c.Query("uuid")
	sessionId := c.Query("sessionid")
	reset := c.Query("reset")

	if number == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "number is required"})
		return
	}

	// check if user exists
	user := utils.GetUserByUsername(number)
	if user.Id == 0 {
		// get unit
		unit := utils.GetUnitByUuid(uuid)

		// generate code
		code := utils.GenerateCode(6)

		// send sms with code
		status := utils.SendSMSCode(number, code, unit, "digest="+digest+"&uuid="+uuid+"&sessionid="+sessionId)

		// check response
		if status != 201 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "authorization code not send"})
			return
		}

		// create user
		days := utils.GetHotspotPreferencesByKey(unit.HotspotId, "user_expiration_days")
		daysInt, _ := strconv.Atoi(days.Value)

		down := utils.GetHotspotPreferencesByKey(unit.HotspotId, "CoovaChilli-Bandwidth-Max-Down")
		downInt, _ := strconv.Atoi(down.Value)
		up := utils.GetHotspotPreferencesByKey(unit.HotspotId, "CoovaChilli-Bandwidth-Max-Up")
		upInt, _ := strconv.Atoi(up.Value)

		autoLogin := utils.GetHotspotPreferencesByKey(unit.HotspotId, "auto_login")
		autoLoginBool, _ := strconv.ParseBool(autoLogin.Value)

		newUser := models.User{
			HotspotId:   unit.HotspotId,
			Name:        number,
			Username:    number,
			Password:    code,
			Email:       "",
			AccountType: "sms",
			KbpsDown:    downInt,
			KbpsUp:      upInt,
			AutoLogin:   autoLoginBool,
			ValidFrom:   time.Now().UTC(),
			ValidUntil:  time.Now().UTC().AddDate(0, 0, daysInt),
		}
		newUser.Id = methods.CreateUser(newUser)

		// create marketing info with user infos
		utils.CreateUserMarketing(newUser.Id, smsMarketingData{Number: number}, "sms")

		// response to client
		c.JSON(http.StatusOK, gin.H{"user_id": number})
	} else {
		// update user info
		days := utils.GetHotspotPreferencesByKey(user.HotspotId, "user_expiration_days")
		daysInt, _ := strconv.Atoi(days.Value)
		user.ValidUntil = time.Now().UTC().AddDate(0, 0, daysInt)

		// create user session check
		utils.CreateUserSession(user.Id, sessionId)

		// create marketing info with user infos
		utils.CreateUserMarketing(user.Id, smsMarketingData{Number: number}, "sms")

		// check if is reset
		if reset == "true" {
			// get unit
			unit := utils.GetUnitByUuid(uuid)

			// generate code
			code := utils.GenerateCode(6)

			// send sms with code
			status := utils.SendSMSCode(number, code, unit, "digest="+digest+"&uuid="+uuid+"&sessionid="+sessionId)
			// check response
			if status != 201 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "authorization code not send"})
				return
			}

			// update code
			user.Password = code
		}

		db := database.Database()
		db.Save(&user)
		db.Close()

		// response to client
		c.JSON(http.StatusOK, gin.H{"user_id": number, "exists": true, "reset": reset, "user_db_id": user.Id})
	}
}

func EmailAuth(c *gin.Context) {
	email := c.Param("email")
	digest := c.Query("digest")
	uuid := c.Query("uuid")
	sessionId := c.Query("sessionid")
	reset := c.Query("reset")

	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "email is required"})
		return
	}

	// check if user exists
	user := utils.GetUserByUsername(email)
	if user.Id == 0 {
		// get unit
		unit := utils.GetUnitByUuid(uuid)

		// generate code
		code := utils.GenerateCode(6)

		// send email with code
		status := utils.SendEmailCode(email, code, unit, "digest="+digest+"&uuid="+uuid+"&sessionid="+sessionId)

		// check response
		if !status {
			c.JSON(http.StatusBadRequest, gin.H{"error": "authorization code not sent"})
			return
		}

		// create user
		days := utils.GetHotspotPreferencesByKey(unit.HotspotId, "user_expiration_days")
		daysInt, _ := strconv.Atoi(days.Value)

		down := utils.GetHotspotPreferencesByKey(unit.HotspotId, "CoovaChilli-Bandwidth-Max-Down")
		downInt, _ := strconv.Atoi(down.Value)
		up := utils.GetHotspotPreferencesByKey(unit.HotspotId, "CoovaChilli-Bandwidth-Max-Up")
		upInt, _ := strconv.Atoi(up.Value)

		autoLogin := utils.GetHotspotPreferencesByKey(unit.HotspotId, "auto_login")
		autoLoginBool, _ := strconv.ParseBool(autoLogin.Value)

		newUser := models.User{
			HotspotId:   unit.HotspotId,
			Name:        email,
			Username:    email,
			Password:    code,
			Email:       email,
			AccountType: "email",
			KbpsDown:    downInt,
			KbpsUp:      upInt,
			AutoLogin:   autoLoginBool,
			ValidFrom:   time.Now().UTC(),
			ValidUntil:  time.Now().UTC().AddDate(0, 0, daysInt),
		}
		newUser.Id = methods.CreateUser(newUser)

		// create marketing info with user infos
		utils.CreateUserMarketing(newUser.Id, emailMarketingData{Email: email}, "email")

		// response to client
		c.JSON(http.StatusOK, gin.H{"user_id": email})
	} else {
		// update user info
		days := utils.GetHotspotPreferencesByKey(user.HotspotId, "user_expiration_days")
		daysInt, _ := strconv.Atoi(days.Value)
		user.ValidUntil = time.Now().UTC().AddDate(0, 0, daysInt)

		// create user session check
		utils.CreateUserSession(user.Id, sessionId)

		// create marketing info with user infos
		utils.CreateUserMarketing(user.Id, emailMarketingData{Email: email}, "email")

		// check if is reset
		if reset == "true" {
			// get unit
			unit := utils.GetUnitByUuid(uuid)

			// generate code
			code := utils.GenerateCode(6)

			// send email with code
			status := utils.SendEmailCode(email, code, unit, "digest="+digest+"&uuid="+uuid+"&sessionid="+sessionId)

			// check response
			if !status {
				c.JSON(http.StatusBadRequest, gin.H{"error": "authorization code not send"})
				return
			}

			// update code
			user.Password = code
		}

		db := database.Database()
		db.Save(&user)
		db.Close()

		// response to client
		c.JSON(http.StatusOK, gin.H{"user_id": email, "exists": true, "reset": reset, "user_db_id": user.Id})
	}
}

func VoucherAuth(c *gin.Context) {
	code := c.Param("code")
	uuid := c.Query("uuid")

	// extract unit
	unit := utils.GetUnitByUuid(uuid)

	// extract voucher
	voucher := utils.GetVoucherByCode(code, unit.HotspotId)

	// check voucher validity
	if voucher.Id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Voucher is invalid"})
	} else {
		if !voucher.Expires.IsZero() && voucher.Expires.Before(time.Now().UTC()) {
			c.JSON(http.StatusOK, gin.H{"message": "Voucher is expired"})
		} else {
			// read hotspot preferences
			days := utils.GetHotspotPreferencesByKey(unit.HotspotId, "voucher_expiration_days")
			daysInt, _ := strconv.Atoi(days.Value)

			// update voucher expiration
			voucher.Expires = time.Now().UTC().AddDate(0, 0, daysInt)

			db := database.Database()
			db.Save(&voucher)
			db.Close()

			c.JSON(http.StatusOK, gin.H{"message": "Voucher is valid"})
		}
	}

}
