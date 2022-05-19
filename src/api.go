package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"strconv"
	"math"
)

func auth(c *gin.Context) {
	if c.GetHeader("Username") == "admin" && c.GetHeader("Password") == "123" {
		token, err := GenerateJWT()
		if err != nil {
			panic(err)
		}

		c.JSON(200, gin.H{
			"Token": token})
	} else {
		c.JSON(401, gin.H{
			"auth_error": "not authorized"})
	}
}

func getUser(c *gin.Context) {
	if !isAuthorized(c.GetHeader("Token")) {
		c.Status(401)
		return
	}

	user := User{}
	user.ID = c.Param("id")
	status := user.GetById()
	if status == DbError {
		c.Status(500)
		return
	} else if status == NotFoundError {
		c.Status(400)
		return
	}

	valid, reason := user.Validate()
	if !valid {
		c.Status(500)
		log.Printf("Invalid user queried from the database (%v) %v", reason, user)
		return
	}

	c.JSON(200, user.ToDisplay())
}

func putUser(c *gin.Context) {
	if !isAuthorized(c.GetHeader("Token")) {
		c.Status(401)
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}

	userDisplay := UserDisplayDTO{}
	err = json.Unmarshal(body, &userDisplay)
	if err != nil {
		c.Status(400)
	}

	userDisplay.ID = c.Param("id")

	valid, desc := userDisplay.Validate()
	if !valid {
		c.JSON(400, gin.H{
			"invalid_reason": desc,
		})
		return
	}

	queryUser := User{ID: userDisplay.ID}
	status := queryUser.GetById()
	if status == DbError {
		c.Status(500)
	} else if status == NotFoundError {
		c.Status(404)
	}
	combinedUser := userDisplay.CombineWithUser(queryUser)

	status = combinedUser.Update()
	if status == DbError {
		c.Status(500)
		return
	} else if status == NotFoundError {
		c.Status(404)
		return
	}
}

func patchUser(c *gin.Context) {
	if !isAuthorized(c.GetHeader("Token")) {
		c.Status(401)
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}

	userDisplay := UserDisplayDTO{}
	err = json.Unmarshal(body, &userDisplay)
	if err != nil {
		c.Status(400)
	}
	userDisplay.ID = c.Param("id")

	valid, desc := userDisplay.ValidateButIgnoreZeroed()
	if !valid {
		c.JSON(400, gin.H{
			"invalid_reason": desc,
		})
		return
	}

	user := userDisplay.ToUser()
	status := user.UpdateSelective()
	if status == DbError {
		c.Status(500)
		return
	} else if status == NotFoundError {
		c.Status(404)
		return
	}
}

func postUser(c *gin.Context) {
	if !isAuthorized(c.GetHeader("Token")) {
		c.Status(401)
		return
	}

	userCreation := UserCreationDTO{}
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(body, &userCreation)
	if err != nil {
		c.Status(400)
	}

	valid, desc := userCreation.Validate()
	if !valid {
		c.JSON(400, gin.H{
			"invalid_reason": desc,
		})

		return
	}

	user := userCreation.ToUser()
	user.ID = uuid.NewString()
	user.RandomKey = randomString()

	status := user.Create()
	if status == DbError {
		c.Status(500)
		return
	} else if status == NotFoundError {
		c.Status(404)
		return
	}

	c.JSON(200, gin.H{
		"id": user.ID,
	})
}

// /api/v1/users?page=<value>&order_by=<[id | firstname | secondname | age | lon | lat]>&filter_by=<[id | firstname | secondname | age | lon | lat]>
// .. &filter_pred=[g | l | ge | le | e | ne]&filter_value=<string>
func getUserPage(c *gin.Context) {
	const perPageItemsCount = 5
	if !isAuthorized(c.GetHeader("Token")) {
		c.Status(401)
		return
	}

	_page := c.Query("page")
	page, err := strconv.Atoi(_page)
	if err != nil {
		c.Status(404)
		return
	}

	filterBy := c.Query("filter_by")
	filterPred := c.Query("filter_pred")
	filterValue := c.Query("filter_value")
	orderBy := c.Query("order_by")

	var filter bool
	var comparisonOperator string
	if filterBy != "" {
		filter = true

		// Note: such kind of conversions will break if names change. But it will work for now
		if filterBy == "lon" {
			filterBy = "x"
		}
		if filterBy == "lat" {
			filterBy = "y"
		}

		if filterPred == "" {
			c.Status(400)
			return
		}

		switch filterPred {
		case "ge":
			comparisonOperator = ">="
		case "le":
			comparisonOperator = "<="
		case "e":
			comparisonOperator = "="
		case "ne":
			comparisonOperator = "<>"
		case "g":
			comparisonOperator = ">"
		case "l":
			comparisonOperator = "<"
		default:
			c.Status(400)
			return
		}
	}

	var queryFind string
	var users []User
	if filter {
		queryFind = filterBy + comparisonOperator + filterValue
	}

	if orderBy != "" {
		// Note: same as above
		if orderBy == "lon" {
			orderBy = "x"
		}
		if orderBy == "lat" {
			orderBy = "y"
		}
	} else {
		orderBy = "id"
	}

	result := db.Find(&users, queryFind)
	if result.Error != nil {
		c.Status(404)
	}

	itemsTotal := result.RowsAffected
	pagesTotal := int(math.Ceil(float64(itemsTotal / perPageItemsCount)))

	if page >= pagesTotal || page < 0 {
		c.JSON(404, gin.H { 
			"error": "You have requested an invalid page",
			"pages_total": pagesTotal,
		})
	}

	err = db.Limit(perPageItemsCount).Offset(page * perPageItemsCount).Order(orderBy).Find(&users, queryFind).Error
	if err != nil {
		c.Status(404)
	}

	usersDisplayDtos := make([]UserDisplayDTO, len(users))
	for i, user := range users {
		valid, reason := user.Validate()
		if !valid {
			// If some object has an invalid state I decide to not send anything. Though it can be changed.
			log.Printf("An invalid object has been encountered while processing a request: %v\n (%v)", reason, user)
			c.Status(500)
			return
		}
		usersDisplayDtos[i] = user.ToDisplay()
	}

	c.JSON(200, gin.H{
		"page_current": page,
		"pages_total": pagesTotal,
		"page_items_count": perPageItemsCount,
		"page_items_total": itemsTotal,
		"items": usersDisplayDtos,
	})
}

func initializeServer() {
	r := gin.Default()

	r.GET("/api/v0/auth", auth)
	r.GET("/api/v1/users/:id", getUser)
	r.GET("/api/v1/users", getUserPage)
	r.PUT("/api/v1/users/:id", putUser)
	r.PATCH("/api/v1/users/:id", patchUser)
	r.POST("/api/v1/users", postUser)
	r.Run()
}
