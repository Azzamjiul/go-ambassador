package controllers

import (
	"ambassador/src/database"
	"ambassador/src/models"
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Products(c *fiber.Ctx) error {
	var products []models.Product
	database.DB.Find(&products)
	return c.JSON(products)
}

func CreateProducts(c *fiber.Ctx) error {
	var product models.Product
	if err := c.BodyParser(&product); err != nil {
		return err
	}
	database.DB.Create(&product)
	return c.JSON(product)
}

func GetProduct(c *fiber.Ctx) error {
	var product models.Product
	id, _ := strconv.Atoi(c.Params("id"))
	product.Id = uint(id)
	database.DB.Find(&product)
	return c.JSON(product)
}

func UpdateProduct(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	product := models.Product{}
	product.Id = uint(id)
	if err := c.BodyParser(&product); err != nil {
		return err
	}
	database.DB.Model(&product).Updates(&product)
	go DeleteCache("products_frontend")
	go DeleteCache("products_backend")
	return c.JSON(product)
}

func DeleteCache(key string) {
	time.Sleep(5 * time.Second)
	database.Cache.Del(context.Background(), key)
}

func DeleteProduct(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	product := models.Product{}
	product.Id = uint(id)
	database.DB.Delete(&product)
	return nil
}

func ProductsFrontend(c *fiber.Ctx) error {
	var products []models.Product
	var ctx = context.Background()

	result, err := database.Cache.Get(ctx, "products_frontend").Result()
	if err != nil {
		database.DB.Find(&products)

		bytes, err := json.Marshal(products)
		if err != nil {
			panic(err)
		}

		if errKey := database.Cache.Set(ctx, "products_frontend", bytes, 30*time.Minute).Err(); errKey != nil {
			panic(errKey)
		}
	} else {
		json.Unmarshal([]byte(result), &products)
	}
	return c.JSON(products)
}

func ProductsBackend(c *fiber.Ctx) error {
	var products []models.Product
	var ctx = context.Background()

	result, err := database.Cache.Get(ctx, "products_backend").Result()
	if err != nil {
		database.DB.Find(&products)

		bytes, err := json.Marshal(products)
		if err != nil {
			panic(err)
		}

		database.Cache.Set(ctx, "products_backend", bytes, 30*time.Minute).Err()
	} else {
		json.Unmarshal([]byte(result), &products)
	}

	var searchedProdcuts []models.Product

	if s := c.Query("s"); s != "" {
		lower := strings.ToLower(s)
		for _, product := range products {
			if strings.Contains(strings.ToLower(product.Title), lower) || strings.Contains(strings.ToLower(product.Description), lower) {
				searchedProdcuts = append(searchedProdcuts, product)
			}
		}
	} else {
		searchedProdcuts = products
	}

	if sortParam := c.Query("sort"); sortParam != "" {
		sortLower := strings.ToLower(sortParam)
		if sortLower == "asc" {
			sort.Slice(searchedProdcuts, func(i, j int) bool {
				return searchedProdcuts[i].Price < searchedProdcuts[j].Price
			})
		} else if sortLower == "desc" {
			sort.Slice(searchedProdcuts, func(i, j int) bool {
				return searchedProdcuts[i].Price > searchedProdcuts[j].Price
			})
		}
	}

	total := len(searchedProdcuts)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	perPage := 9

	var data []models.Product
	if total <= page*perPage && total >= (page-1)*perPage {
		data = searchedProdcuts[(page-1)*perPage : total]
	} else if total >= page*perPage {
		data = searchedProdcuts[(page-1)*perPage : page*perPage]
	} else {
		data = []models.Product{}
	}

	return c.JSON(fiber.Map{
		"total":     total,
		"page":      page,
		"last_page": total/perPage + 1,
		"data":      data,
	})
}
