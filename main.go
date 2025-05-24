package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"encoding/json"

	"main/config"
	"main/tasks"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/sirupsen/logrus"

	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"context"
	"strings"

	"github.com/gin-contrib/cors"

	_ "main/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var db *gorm.DB

// connect to DB postgreSQL
func initDB() {
	cfg := config.Load()
	fmt.Printf("Port:", cfg.DBPort)
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)
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
	db.AutoMigrate(&User{})

	log.Printf("База запущена успешно!")
}

// JWT TOKEN
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
	expirationTime := time.Now().Add(30 * time.Minute)      // Время жизни access токена
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

// @Summary		login
// @Description	Logs in a user and returns access and refresh tokens
// @Tags			auth
// @Accept			json
// @Produce		json
// @Security		TokenAuth
// @Param			credentials	body		Credentials			true	"Данные для входа"
// @Success		200			{object}	map[string]string	"{"accesstoken": "string", "refreshtoken": "string"}"
// @Failure		400			{object}	map[string]string	"{"message": "string"}"
// @Failure		401			{object}	map[string]string	"{"message": "string"}"
// @Router			/login [post]
func login(c *gin.Context) {
	var creds Credentials
	if err := c.BindJSON(&creds); err != nil {
		//c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		handleError(c, http.StatusBadRequest, "invalid request")
		return
	}

	// Ищем пользователя по username
	var user User
	if err := db.Where("login = ?", creds.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			handleError(c, http.StatusUnauthorized, "User not found")
		} else {
			handleError(c, http.StatusInternalServerError, "Database error")
		}
		return
	}

	// Сравниваем хешированный пароль
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password))
	if err != nil {
		handleError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Генерация токенов
	accesstoken, refreshtoken, err := generateToken(creds.Username)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "could not create token")
		return
	}

	//handleError(c, http.StatusOK, "invalid request")
	c.JSON(http.StatusOK, gin.H{"accesstoken": accesstoken, "refreshtoken": refreshtoken})
}

// Register godoc
// @Summary      Регистрация нового пользователя
// @Description  Регистрирует нового пользователя с уникальным логином и хешированным паролем
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     TokenAuth
// @Param        user  body      Credentials  true  "Данные для регистрации (логин и пароль)"
// @Success      201   {object}  map[string]string  "Пользователь успешно зарегистрирован"
// @Failure      400   {object}  map[string]string  "Некорректный формат запроса"
// @Failure      409   {object}  map[string]string  "Пользователь уже существует"
// @Failure      500   {object}  map[string]string  "Внутренняя ошибка сервера"
// @Router       /register [post]
func register(c *gin.Context) {
	var creds Credentials
	var user User
	if err := c.BindJSON(&creds); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request")
		return
	}

	if err := db.Where("login = ?", creds.Username).First(&user).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			handleError(c, http.StatusUnauthorized, "User already exists")
			return
		}
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	newUser := User{
		Login:      creds.Username,
		Password:   string(hashedPassword),
		UsertypeID: "UT01",
	}
	if err := db.Create(&newUser).Error; err != nil {
		handleError(c, http.StatusInternalServerError, "Login already exists")
		return
	}

	// Можно сразу создать токены или просто сообщить об успешной регистрации
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
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
	//Itemid        string  `gorm:"primaryKey;column:item_id" json:"item_id"`
	Name          string `json:"name"`
	Category_id   string `json:"category_id"`
	Description   string `json:"description"`
	Price         int    `json:"price"`
	QuantityStock int32  `json:"quantitystock"`
}

type OrderItem struct {
	//OrderItemid   string  `gorm:"primaryKey;column:orderitem_id" json:"orderitem_id"`
	OrderId         int     `json:"order_id"`
	Itemid          int     `json:"item_id"`
	Quantity        int     `json:"quantity"`
	SumPriceItem    int     `json:"sum_price_item"`
	PackageId       string  `json:"package_id"`
	QuantityPackage int     `json:"quantity_package"`
	SumPricePackage float64 `json:"sum_price_package"`
	UserId          int     `json:"user_id"`
}

type User struct {
	//UserID     string `gorm:"primaryKey;column:user_id" json:"user_id"`
	Name       string `gorm:"column:name" json:"name"`
	Lastname   string `gorm:"column:lastname" json:"lastname"`
	Surname    string `gorm:"column:surname" json:"surname"`
	Email      string `gorm:"column:email" json:"email"`
	Phone      string `gorm:"column:phone" json:"phone"`
	UsertypeID string `gorm:"column:usertype_id" json:"usertype_id"`
	Login      string `gorm:"column:login" json:"login"`
	Password   string `gorm:"column:password" json:"password"`
}

type Order struct {
	OrderID         int     `gorm:"primaryKey;column:order_id" json:"order_id"`
	UserID          int     `gorm:"column:user_id" json:"user_id"`
	OrderDate       string  `gorm:"column:orderdate" json:"order_date"` // можно time.Time, если используешь time пакет
	DeliveryTypeID  string  `gorm:"column:deliverytype_id" json:"delivery_type_id"`
	AddressDelivery string  `gorm:"column:addressdelivery" json:"address_delivery"`
	NumberKilometer float64 `gorm:"column:numberkilometer" json:"number_kilometer"`
	DeliveryPrice   float64 `gorm:"column:deliveryprice" json:"delivery_price"`
	//TimeDelivery    string  `gorm:"column:timedelivery" json:"time_delivery"` // или time.Time если используешь TIME тип
	FinalPrice    float64 `gorm:"column:finalprice" json:"final_price"`
	OrderStatusID string  `gorm:"column:orderstatus_id" json:"order_status_id"`
}

// ORDER

// getOrder godoc
// @Summary      Получить все заказы пользователя
// @Description  Возвращает список заказов по user_id
// @Tags         order
// @Accept       json
// @Produce      json
// @Security	 TokenAuth
// @Param        user_id  query     integer  true  "ID пользователя"
// @Success      200      {array}   Order
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /order [get]
func getOrder(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User_id is required"})
		return
	}

	var orders []Order
	if err := db.Debug().Where("user_id = ?", userID).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

type CreateOrderRequest struct {
	OrderItems []int `json:"order_items" example:"[1, 2, 3]"`
	Order      Order `json:"orderr"`
}


// @Summary      Create order
// @Description  Add a new order with item details and user ID
// @Tags         order
// @Accept       json
// @Produce      json
// @Security     TokenAuth
// @Param request body CreateOrderRequest true "Order creation payload"
// @Success      201 {object} Order "Order created successfully"
// @Failure      400 {object} map[string]string "Invalid request"
// @Failure      500 {object} map[string]string "Failed to create order"
// @Router       /order [post]
func createOrder(c *gin.Context) {
	var requestData map[string]interface{}

	// Прочитать тело запроса в map
	if err := c.BindJSON(&requestData); err != nil {
		handleError(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Получить массив order_items
	rawItems, ok := requestData["order_items"].([]interface{})
	if !ok || len(rawItems) == 0 {
		handleError(c, http.StatusBadRequest, "order_items is required and must be a non-empty array")
		return
	}

	// Преобразуем []interface{} в []int
	orderItemIDs := make([]int, len(rawItems))
	for i, val := range rawItems {
		floatVal, ok := val.(float64) // JSON числа приходят как float64
		if !ok {
			handleError(c, http.StatusBadRequest, "Invalid order_items format")
			return
		}
		orderItemIDs[i] = int(floatVal)
	}

	// Получить item как map и преобразовать в структуру Item
	itemData, ok := requestData["orderr"].(map[string]interface{})
	if !ok {
		handleError(c, http.StatusBadRequest, "Order data is required")
		return
	}

	// Преобразовать itemData в Item
	var newOrder Order
	itemBytes, _ := json.Marshal(itemData)
	if err := json.Unmarshal(itemBytes, &newOrder); err != nil {
		handleError(c, http.StatusBadRequest, "Invalid item format")
		return
	}

	if err := db.Create(&newOrder).Error; err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to create order")
		return
	}
	// Здесь newOrder.ID уже содержит номер созданного заказа
	c.JSON(http.StatusCreated, gin.H{
		"message":  "Order created successfully",
		"order_id": newOrder.OrderID,
	})

	if err := db.Model(&OrderItem{}).Where("orderitem_id IN ?", orderItemIDs).Update("order_id", newOrder.OrderID).Error; err != nil {
		log.Printf("Ошибка при обновлении order_items: %v", err)
	}

	//db.Create(&newOrder)

	c.JSON(http.StatusCreated, newOrder)
}

// @Summary			Delete Order
// @Description	Remove a order from the catalog by its ID
// @Tags			order
// @Accept			json
// @Produce			json
// @Security		TokenAuth
// @Param			order_id	query		integer				true	"Order ID"
// @Success			200	{object}	map[string]string	"Message indicating successful deletion"
// @Failure			404	{object}	map[string]string	"Item not found"
// @Router			/order [delete]
func deleteOrder(c *gin.Context) {
	order_idSTR := c.Query("order_id")

	// Преобразуем в int
	order_id, err := strconv.Atoi(order_idSTR)
	if err != nil {
		handleError(c, http.StatusBadRequest, "invalid user_id")
		return
	}

	if err := db.Where("order_id = ?", order_id).Delete(&OrderItem{}).Error; err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to delete orderitems")
		return
	}

	result := db.Delete(&Order{}, "order_id = ?", order_id)

	if result.Error != nil {
		handleError(c, http.StatusInternalServerError, "Failed to delete order")
		return
	}

	if result.RowsAffected == 0 {
		handleError(c, http.StatusNotFound, "Order not found")
		return
	}

	handleError(c, http.StatusOK, "Order deleted")

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

// @Summary		getItems
// @Description	Get a list of products with pagination, sorting, and filtering options
// @Tags			items
// @Accept			json
// @Produce		json
// @Security		TokenAuth
// @Param			page		query		int		false	"Page number"		(default: 1)
// @Param			limit		query		int		false	"Limit per page"	(default: 10)
// @Param			name		query		string	false	"Product name"
// @Param			category_id	query		string	false	"Product category"
// @Param			sort		query		string	false	"Sort by field (e.g., field_name:asc or field_name:desc)"
// @Success		200			{object}	map[string]interface{}
// @Failure		400			{object}	map[string]interface{}	"Bad Request"
// @Failure		500			{object}	map[string]interface{}	"Internal Server Error"
// @Router			/items [get]
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
	c.JSON(http.StatusOK, gin.H{
		"data":  items,
		"total": total,
		"page":  pageInt,
		"limit": limitInt,
	})
}

// getItemsBasket godoc
// @Summary      Получить все позиции корзины пользователя
// @Description  Возвращает список товаров в корзине по user_id
// @Tags         basket
// @Accept       json
// @Produce      json
// @Security	 TokenAuth
// @Param        user_id  query     integer  true  "ID пользователя"
// @Success      200      {array}   OrderItem
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /basket/items [get]
func getItemsBasket(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	var itemsBasket []OrderItem
	if err := db.Debug().Where("user_id = ?", userID).Find(&itemsBasket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, itemsBasket)
}

// @Summary		Get Item by ID
// @Description	Get a single item by its ID
// @Tags			items
// @Accept			json
// @Produce		json
// @Security		TokenAuth
// @Param			id	path		string	true	"Item ID"
// @Success		200	{object}	Item
// @Failure		404	{object}	map[string]string
// @Router			/items/{id} [get]
func getItemByID(c *gin.Context) {
	id := string(c.Param("id"))
	var item Item
	if err := db.First(&item, "item_id = ?", id).Error; err != nil {
		handleError(c, http.StatusNotFound, "item not found "+id)
		//c.JSON(http.StatusNotFound, gin.H{"message": "item not found " + id})
		return
	}
	c.JSON(http.StatusOK, item)
}

// @Summary		Create item
// @Description	Add a new item to the stock
// @Tags			items
// @Accept			json
// @Produce		json
// @Security		TokenAuth
// @Param			product	body		Item				true	"New item details"
// @Success		201		{object}	Item				"Item created successfully"
// @Failure		400		{object}	map[string]string	"Invalid request"
// @Router			/items [post]
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

// @Summary		Add To Basket
// @Description	Add a new item to the user's basket by user_id and item_id
// @Tags			basket
// @Accept			json
// @Produce		json
// @Security		TokenAuth
// @Param			user_id1		query		integer				true	"User ID"
// @Param			item_id1		query		integer				true	"Item ID"
// @Param			item		body		OrderItem			true	"Basket item details (e.g., quantity)"
// @Success		201			{object}	OrderItem			"Item added to basket successfully"
// @Failure		400			{object}	map[string]string	"Invalid request"
// @Failure		500			{object}	map[string]string	"Failed to add item to basket"
// @Router			/basket [post]
func createItemBasket(c *gin.Context) {
	var newItem OrderItem

	// Получаем user_id и item_id из URL-параметров
	userIDStr := c.Query("user_id1")
	itemIDStr := c.Query("item_id1")

	if err := c.BindJSON(&newItem); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request")
		return
	}

	// Преобразуем в int
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		handleError(c, http.StatusBadRequest, "invalid user_id")
		return
	}

	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		handleError(c, http.StatusBadRequest, "invalid item_id")
		return
	}

	// Устанавливаем user_id и item_id из параметров в структуру
	newItem.UserId = userID
	newItem.Itemid = itemID

	if err := db.Create(&newItem).Error; err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to add item to basket")
		return
	}

	c.JSON(http.StatusCreated, "Access to add item to basket")
}

// @Summary		Обновление товара
// @Description	Update an existing item by its ID
// @Tags			items
// @Accept			json
// @Produce		json
// @Security		TokenAuth
// @Param			id		path		string				true	"Item ID"
// @Param			product	body		Item				true	"Updated item"
// @Success		200		{object}	Item				"Updated item details"
// @Failure		400		{object}	map[string]string	"Invalid request"
// @Failure		404		{object}	map[string]string	"Item not found"
// @Router			/items/{id} [put]
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

// @Summary		Delete Item
// @Description	Remove a item from the catalog by its ID
// @Tags			items
// @Accept			json
// @Produce		json
// @Security		TokenAuth
// @Param			id	path		string				true	"Item ID"
// @Success		200	{object}	map[string]string	"Message indicating successful deletion"
// @Failure		404	{object}	map[string]string	"Item not found"
// @Router			/items/{id} [delete]
func deleteItem(c *gin.Context) {
	id := c.Param("id")

	if err := db.Delete(&Item{}, "item_id = ?", id).Error; err != nil {
		handleError(c, http.StatusNotFound, "item not found")
		return
	} else {
		handleError(c, http.StatusOK, "item deleted")
	}
}

// @Summary		Delete From Basket
// @Description	Remove a item from the user's basket
// @Tags			basket
// @Accept			json
// @Produce		json
// @Security		TokenAuth
// @Param			item_id	query		integer	true	"Item ID"
// @Param			user_id	query		integer	true	"User ID"
// @Success		200	{object}	map[string]string
// @Failure		404	{object}	map[string]string
// @Router			/basket [delete]
func deleteItemBasket(c *gin.Context) {

	// Получаем user_id и item_id из URL-параметров
	UserId := c.Query("user_id")
	ItemId := c.Query("item_id")
	//ItemId := c.Param("id")

	//var item OrderItem
	result := db.Where("itemid = ? AND user_id = ?", ItemId, UserId).Delete(&OrderItem{})
	if result.RowsAffected == 0 {
		handleError(c, http.StatusNotFound, "Item not found in basket")
		return
	}
	if result.Error != nil {
		handleError(c, http.StatusInternalServerError, "Failed to remove item from basket")
		return
	}

	handleError(c, http.StatusOK, "Item removed from basket")
}

// @title						Bakery API
// @version					1.0
// @description				Это API для интернет-магазина мягких игрушек
// @host						localhost:8080
// @BasePath					/
// @securityDefinitions.apikey	TokenAuth
// @in							header
// @name						Authorization
// @description				Введите ваш токен напрямую в заголовке Authorization
func main() {

	// Создаем новый логгер
	log := logrus.New()

	// Устанавливаем формат вывода (JSON или текст — на выбор)
	log.SetFormatter(&logrus.JSONFormatter{}) // Можно заменить на &logrus.TextFormatter{}

	// Устанавливаем уровень логирования
	log.SetLevel(logrus.InfoLevel) // Будут выводиться Info, Warn, Error и выше

	// Пытаемся открыть файл для записи логов
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(file) // Если файл открыт, пишем логи в него
	} else {
		log.Warn("Не удалось открыть файл логов, логирование в консоль")
	}

	// Примеры логов разных уровней
	log.Info("Информационное сообщение")
	log.Warn("Предупреждение")
	log.Error("Ошибка")
	// log.Debug("Отладочная информация") — не будет выведено, если уровень Info

	//router := gin.Default()

	initDB()
	time.Sleep(10 * time.Second)
	var router = gin.Default()
	router.Use(cors.A()) // разрешает все CORS-запросы

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	protected := router.Group("/")

	router.POST("/login", login)
	router.POST("/register", register)
	protected.Use(authMiddleware())
	{

		protected.GET("/items", getItems)

		protected.GET("/itemsWT", getItemsWithTimeout)

		// Получение всех товаров из корзины
		protected.GET("/basket/items", getItemsBasket)

		// Получение товара по ID
		protected.GET("/items/:id", getItemByID)

		// Создание нового товара
		protected.POST("/items", createItem)

		// Добавление товара в корзину
		protected.POST("/basket", createItemBasket)

		// Обновление существующего товара
		protected.PUT("/items/:id", updateItem)

		// Удаление товара
		protected.DELETE("/items/:id", deleteItem)

		// Удаление товара из корзины
		protected.DELETE("/basket", deleteItemBasket)

		// Получение всех заказов
		protected.GET("/order", getOrder)

		//Создание заказа
		protected.POST("/order", createOrder)

		//Удаление заказа
		protected.DELETE("/order",deleteOrder)
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
