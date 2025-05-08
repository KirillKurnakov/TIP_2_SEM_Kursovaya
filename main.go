package main

import (
	"net/http"
	"strconv"
	"time"
	"os"
	"fmt"

	"main/tasks"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"

	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"context"
	"strings"

	_ "main/docs"

	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"

)

var db *gorm.DB

// connect to DB postgreSQL
func initDB() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	);
	//dsn := os.Getenv("DATABASE_URL");
	//dsn := "host=localhost user=postgres password=postgres dbname=tipOnlineShop port=5432 sslmode=disable"
	//dsn := "host=db user=postgres password=postgres dbname=tipOnlineShop port=5432 sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Миграция схемы
	db.AutoMigrate(&Item{})
	db.AutoMigrate(&OrderItem{})
}


//JWT TOKEN 
	var jwtKey = []byte("my_secret_key")

        type Credentials struct {
            Username string `json:"username"`
            Password string `json:"password"`
        }

        type Claims struct {
            Username string `json:"username"`
            jwt.StandardClaims
        }

		var users = []Credentials{
			{Username: "user", Password: "password"},
			{Username: "user1", Password: "password1"},
			{Username: "user2", Password: "password2"},
			{Username: "user3", Password: "password3"},
		}
		func handleError(c *gin.Context, statusCode int, message string) {
			c.JSON(statusCode, gin.H{"error": message})
		}

		// Генерация access и refresh токенов
		func generateToken(username string) (string, string, error) {
			expirationTime := time.Now().Add(30 * time.Minute) // Время жизни access токена
			refreshExpirationTime := time.Now().Add(24 * time.Hour) // Время жизни refresh токена
		
			// Создание access токена
			claims := &Claims{
			Username: username,
			StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			},
			}
		
			accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			accessTokenString, err := accessToken.SignedString(jwtKey)
			if err != nil {
			return "", "", err
			}
		
			// Создание refresh токена
			refreshClaims := &Claims{
			Username: username,
			StandardClaims: jwt.StandardClaims{
			ExpiresAt: refreshExpirationTime.Unix(),
			},
			}
		
			refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
			refreshTokenString, err := refreshToken.SignedString(jwtKey)
			if err != nil {
			return "", "", err
			}
		
			return accessTokenString, refreshTokenString, nil
		}

			//	@Summary		login
			//	@Description	Logs in a user and returns access and refresh tokens
			//	@Tags			auth
			//	@Accept			json
			//	@Produce		json
			//	@Security		TokenAuth
			//	@Param			credentials	body		Credentials			true	"Данные для входа"
			//	@Success		200			{object}	map[string]string	"{"accesstoken": "string", "refreshtoken": "string"}"
			//	@Failure		400			{object}	map[string]string	"{"message": "string"}"
			//	@Failure		401			{object}	map[string]string	"{"message": "string"}"
			//	@Router			/login [post]
		func login(c *gin.Context) {
            var creds Credentials
            if err := c.BindJSON(&creds); err != nil {
                //c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
				handleError(c, http.StatusBadRequest, "invalid request")
                return
            }

            var validUser *Credentials
				for _, user := range users {
					if user.Username == creds.Username && user.Password == creds.Password {
						validUser = &user
						break
					}
				}

			if validUser == nil {
				//c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
				handleError(c, http.StatusUnauthorized, "unauthorized")
				return
			}

            accesstoken, refreshtoken, err := generateToken(creds.Username)
            if err != nil {
                //c.JSON(http.StatusInternalServerError, gin.H{"message": "could not create token"})
                handleError(c, http.StatusInternalServerError, "could not create token")
				return
            }

            //handleError(c, http.StatusOK, "invalid request")
			c.JSON(http.StatusOK, gin.H{"accesstoken": accesstoken, "refreshtoken": refreshtoken})
        }

        func authMiddleware() gin.HandlerFunc {
            return func(c *gin.Context) {
                tokenString := c.GetHeader("Authorization")

                claims := &Claims{}
                token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
                    return jwtKey, nil
                })

                if err != nil || !token.Valid {
					if ve, ok := err.(*jwt.ValidationError); ok {
						if ve.Errors&jwt.ValidationErrorExpired != 0 {
							handleError(c, http.StatusUnauthorized, "token expired")
							//c.JSON(http.StatusUnauthorized, gin.H{"message": "token expired"})
							c.Abort()
							return
						}
					}
                    handleError(c, http.StatusUnauthorized, "unauthorized")
					//c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
                    c.Abort()
                    return
                }

                c.Next()
            }
        }



type Item struct {
	Itemid        string  `gorm:"primaryKey;column:item_id" json:"item_id"`
	Name          string  `json:"name"`
	Category_id   string  `json:"category_id"`
	Description   string  `json:"description"`
	Price         int 	  `json:"price"`
	QuantityStock int32     `json:"quantitystock"`
}

type OrderItem struct {
	OrderItemid   string  `gorm:"primaryKey;column:orderitem_id" json:"orderitem_id"`
	OrderId 	  string  `json:"order_id"`
	Itemid        string  `json:"item_id"`
	Quantity 	  int	  `json:"quantity"`
	SumPriceItem  int	  `json:"sum_price_item"`
	PackageId	  string  `json:"package_id"`
	QuantityPackage int   `json:"quantity_package"`
	SumPricePackage float64	`json:"sum_price_package"`
}



func getItemsWithTimeout(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	var items []Item
	if err := db.WithContext(ctx).Find(&items).Error; err != nil {
		handleError(c, http.StatusRequestTimeout, "Request timed out")
		return
	}

	c.JSON(http.StatusOK, items)
}

	//	@Summary		getItems
	//	@Description	Get a list of products with pagination, sorting, and filtering options
	//	@Tags			items
	//	@Accept			json
	//	@Produce		json
	//	@Security		TokenAuth
	//	@Param			page		query		int		false	"Page number"		(default: 1)
	//	@Param			limit		query		int		false	"Limit per page"	(default: 10)
	//	@Param			name		query		string	false	"Product name"
	//	@Param			category_id	query		string	false	"Product category"
	//	@Param			sort		query		string	false	"Sort by field (e.g., field_name:asc or field_name:desc)"
	//	@Success		200			{object}	map[string]interface{}
	//	@Failure		400			{object}	map[string]interface{}	"Bad Request"
	//	@Failure		500			{object}	map[string]interface{}	"Internal Server Error"
	//	@Router			/items [get]
func getItems(c *gin.Context) {
	var items []Item
	var total int64
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	name := c.Query("name")
	category_id := c.Query("category_id")

	sort := c.Query("sort")


	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	offset := (pageInt - 1) * limitInt
	query := db.Limit(limitInt).Offset(offset)

	if name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
		
	}
	if category_id != "" {
		query = query.Where("category_id ILIKE ?", "%"+category_id+"%")
		//handleError(c, http.StatusOK, category_id)
	}

	// Проверяем параметр сортировки и добавляем его к запросу
    if sort != "" {
        // Предположим, что параметр sort может быть в формате "field_name:order"
        // Например: "title:asc" или "author:desc"
        sortParams := strings.Split(sort, ":")
        if len(sortParams) == 2 {
            field := sortParams[0]
            order := sortParams[1]

            // Добавляем сортировку в зависимости от указанного порядка
            if order == "desc" {
                query = query.Order(field + " DESC")
            } else {
                query = query.Order(field + " ASC")
            }
        } else {
            // Если формат неверный, можно использовать значение по умолчанию
            query = query.Order("title ASC") // Сортировка по умолчанию
        }
    }

	query.Find(&items).Count(&total)

	//db.Limit(limitInt).Offset(offset).Find(&items).Count(&total)

	//db.Find(&items)
	c.JSON(http.StatusOK, gin.H {
		"data": items,
		"total": total,
		"page": pageInt,
		"limit": limitInt,
	})
}

//	@Summary		Get Basket Items
//	@Description	Retrieve a list of items in the basket for the current user
//	@Tags			basket
//	@Accept			json
//	@Produce		json
//	@Security		TokenAuth
//	@Success		200	{array}		OrderItem	"List of items in the basket"
//	@Failure		500	{object}	string		"Internal Server Error"
//	@Router			/id/1/basket [get]
func getItemsBasket(c *gin.Context) {
	var itemsBasket []OrderItem
	db.Find(&itemsBasket)
	c.JSON(http.StatusOK, itemsBasket)
}

//	@Summary		Get Item by ID
//	@Description	Get a single item by its ID
//	@Tags			items
//	@Accept			json
//	@Produce		json
//	@Security		TokenAuth
//	@Param			id	path		string	true	"Item ID"
//	@Success		200	{object}	Item
//	@Failure		404	{object}	map[string]string
//	@Router			/items/{id} [get]
func getItemByID(c *gin.Context) {
	id := string(c.Param("id"))
	var item Item
	if err := db.First(&item, "item_id = ?" , id).Error; err != nil {
		handleError(c, http.StatusNotFound, "item not found " + id)
		//c.JSON(http.StatusNotFound, gin.H{"message": "item not found " + id})
		return
	}
	c.JSON(http.StatusOK, item)
}

//	@Summary		Create item
//	@Description	Add a new item to the stock
//	@Tags			items
//	@Accept			json
//	@Produce		json
//	@Security		TokenAuth
//	@Param			product	body		Item				true	"New item details"
//	@Success		201		{object}	Item				"Item created successfully"
//	@Failure		400		{object}	map[string]string	"Invalid request"
//	@Router			/items [post]
func createItem(c *gin.Context) {
	var newItem Item

	if err := c.BindJSON(&newItem); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request")
		//c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	db.Create(&newItem)
	c.JSON(http.StatusCreated, newItem)
}

//	@Summary		Add To Basket
//	@Description	Add a new item to the user's basket
//	@Tags			basket
//	@Accept			json
//	@Produce		json
//	@Security		TokenAuth
//	@Param			item	body		OrderItem			true	"Basket item"
//	@Success		201		{object}	OrderItem			"Item added to basket successfully"
//	@Failure		400		{object}	map[string]string	"Invalid request"
//	@Failure		500		{object}	map[string]string	"Failed to add item to basket"
//	@Router			/id/1/basket [post]
func createItemBasket(c *gin.Context) {
	var newItem OrderItem

	if err := c.BindJSON(&newItem); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request")
		//c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	if err := db.Create(&newItem).Error; err != nil {
		handleError(c, http.StatusInternalServerError, "failed to add item to basket")
		//c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to add item to basket"})
		return
	}

	c.JSON(http.StatusCreated, newItem)
}

//	@Summary		Обновление товара
//	@Description	Update an existing item by its ID
//	@Tags			items
//	@Accept			json
//	@Produce		json
//	@Security		TokenAuth
//	@Param			id		path		string				true	"Item ID"
//	@Param			product	body		Item				true	"Updated item"
//	@Success		200		{object}	Item				"Updated item details"
//	@Failure		400		{object}	map[string]string	"Invalid request"
//	@Failure		404		{object}	map[string]string	"Item not found"
//	@Router			/items/{id} [put]
func updateItem(c *gin.Context) {
	id := c.Param("id")
	var updatedItem Item

	if err := c.BindJSON(&updatedItem); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request")
		//c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	if err := db.Model(&Item{}).Where("item_id = ?", id).Updates(updatedItem).Error; err != nil {
		handleError(c, http.StatusNotFound, "item not found")
		//c.JSON(http.StatusNotFound, gin.H{"message": "item not found"})
		return
	}
	c.JSON(http.StatusOK, updatedItem)
}

//	@Summary		Delete Item
//	@Description	Remove a item from the catalog by its ID
//	@Tags			items
//	@Accept			json
//	@Produce		json
//	@Security		TokenAuth
//	@Param			id	path		string				true	"Item ID"
//	@Success		200	{object}	map[string]string	"Message indicating successful deletion"
//	@Failure		404	{object}	map[string]string	"Item not found"
//	@Router			/items/{id} [delete]
func deleteItem(c *gin.Context) {
	id := c.Param("id")

		if err := db.Delete(&Item{}, "item_id = ?" , id).Error; err != nil {
			handleError(c, http.StatusNotFound, "item not found")
			//c.JSON(http.StatusNotFound, gin.H{"message": "item not found"})
			return
		} else {
			handleError(c, http.StatusOK, "item deleted")
			//c.JSON(http.StatusOK, gin.H{"message": "item deleted"})
		}
	//handleError(c, http.StatusNotFound, "item not found")
	//c.JSON(http.StatusNotFound, gin.H{"message": "item not found"})
}

//	@Summary		Delete From Basket
//	@Description	Remove a item from the user's basket
//	@Tags			basket
//	@Accept			json
//	@Produce		json
//	@Security		TokenAuth
//	@Param			id	path		string	true	"Item ID"
//	@Success		200	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Router			/id/1/basket/{id} [delete]
func deleteItemBasket(c *gin.Context) {
	ItemId := c.Param("id")

	var item OrderItem
	if err := db.Where("itemid = ?", ItemId).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			handleError(c, http.StatusNotFound, "item not found in basket")
			//c.JSON(http.StatusNotFound, gin.H{"message": "item not found in basket"})
		} else {
			handleError(c, http.StatusInternalServerError, "failed to fetch basket item")
			//c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to fetch basket item"})
		}
		return
	}

	if err := db.Delete(&item).Error; err != nil {
		handleError(c, http.StatusInternalServerError, "failed to remove item from basket")
		//c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to remove item from basket"})
		return
	}

	handleError(c, http.StatusOK, "item removed from basket")
	//c.JSON(http.StatusOK, gin.H{"message": "item removed from basket"})
}

//	@title						Bakery API
//	@version					1.0
//	@description				Это API для интернет-магазина мягких игрушек
//	@host						localhost:8080
//	@BasePath					/
//	@securityDefinitions.apikey	TokenAuth
//	@in							header
//	@name						Authorization
//	@description				Введите ваш токен напрямую в заголовке Authorization
func main() {
	//router := gin.Default()
	initDB()
	time.Sleep(10 * time.Second)
	var router = gin.Default()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    protected := router.Group("/")

	router.POST("/login", login)
    protected.Use(authMiddleware())
        {

			protected.GET("/items", getItems)

			protected.GET("/itemsWT", getItemsWithTimeout)

			// Получение всех товаров из корзины
			protected.GET("/id/1/basket", getItemsBasket)

			// Получение товара по ID
			protected.GET("/items/:id", getItemByID)

			// Создание нового товара
			protected.POST("/items", createItem)

			// Добавление товара в корзину
			protected.POST("/id/1/basket", createItemBasket)

			// Обновление существующего товара
			protected.PUT("/items/:id", updateItem)

			// Удаление товара
			protected.DELETE("/items/:id", deleteItem)

			// Удаление товара из корзины
			protected.DELETE("/id/1/basket/:id", deleteItemBasket)
        }

			// Создание задачи
		router.POST("/tasks", func(c *gin.Context) {
			taskID := tasks.CreateTask()
			log.Printf("Task creation requested: ID=%s", taskID)
			go tasks.RunTask(taskID)
			c.JSON(201, gin.H{"task_id": taskID})
		})

		// Получение статуса задачи
		router.GET("/tasks/:id", func(c *gin.Context) {
			taskID := c.Param("id")
			log.Printf("Task status requested: ID=%s", taskID)
			task := tasks.GetTask(taskID)
			if task == nil {
				c.JSON(404, gin.H{"error": "Task not found"})
				return
			}
			c.JSON(200, task)
		})

		// Отмена задачи
		router.POST("/tasks/:id/cancel", func(c *gin.Context) {
			taskID := c.Param("id")
			log.Printf("Cancel request received for Task ID: %s", taskID)
			tasks.CancelTask(taskID)
			c.JSON(200, gin.H{"message": "Task cancellation requested", "task_id": taskID})
	})


	router.Run(":8080")
}