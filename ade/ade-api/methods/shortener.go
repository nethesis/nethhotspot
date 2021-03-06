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
 * author: Matteo Valentini <matteo.valentini@nethesis.it>
 */

package methods

import (
	"net/http"

	"github.com/gin-gonic/gin"

	wax_utils "github.com/nethesis/icaro/wax/utils"
)

func GetLongUrl(c *gin.Context) {
	hash := c.Param("hash")

	shortUrl := wax_utils.GetShortUrlByHash(hash)

	if shortUrl.Id > 0 {
		c.Redirect(http.StatusFound, shortUrl.LongUrl)
		return
	} else {
		c.JSON(http.StatusNotFound, gin.H{"message": "Shortener hash not found!"})
		return
	}
}
