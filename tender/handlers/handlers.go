package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"tender/db"
	"tender/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Неверный формат запроса.")
		if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		errorResponse := models.NewErrorResponse("Ошибка при отправке ответа.")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
}

func CreateTenderHandler(w http.ResponseWriter, r *http.Request) {

	var newTenderRequest models.NewTenderRequest
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.NewErrorResponse("Ошибка чтения тела запроса."))
		return
	}

	re := regexp.MustCompile(`[^\x00-\x7F]+`)
	cleanBody := re.ReplaceAllString(string(body), "")

	if err := json.Unmarshal([]byte(cleanBody), &newTenderRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.NewErrorResponse("Ошибка декодирования JSON: " + err.Error()))
		return
	}

	if newTenderRequest.Name == "" || newTenderRequest.Description == "" || newTenderRequest.ServiceType == "" || newTenderRequest.OrganizationID == "" || newTenderRequest.CreatorUsername == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.NewErrorResponse("Поля name, description, organizationId и creatorUsername обязательны. Возможно поле serviceType неправильно заполнено."))
		return
	}
	var tenders models.Tender

	tender := models.Tender{
		ID:             uuid.New(),
		Name:           newTenderRequest.Name,
		Description:    newTenderRequest.Description,
		Status:         models.TENDER_CREATED,
		ServiceType:    newTenderRequest.ServiceType,
		OrganizationID: newTenderRequest.OrganizationID,
		Version:        tenders.Version,
		CreatedAt:      tenders.CreatedAt,
	}

	if err := db.DB.Create(&tender).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.NewErrorResponse("Сервер не готов обрабатывать запросы."))
		return
	}
	tenderResponse := models.Tender{
		ID:          tender.ID,
		Name:        newTenderRequest.Name,
		Description: newTenderRequest.Description,
		Status:      models.TENDER_CREATED,
		ServiceType: newTenderRequest.ServiceType,
		Version:     tender.Version,
		CreatedAt:   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tenderResponse)
}

func GetTendersHandler(w http.ResponseWriter, r *http.Request) {

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	serviceTypeStr := r.URL.Query().Get("service_type")
	limit := 10
	offset := 0

	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			errorResponse := models.NewErrorResponse("Неверный формат параметра limit.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
	}

	if offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			w.WriteHeader(http.StatusBadRequest)
			errorResponse := models.NewErrorResponse("Неверный формат параметра offset.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
	}

	var tenders []models.Tender
	query := db.DB.Model(&models.Tender{})

	if serviceTypeStr != "" {
		serviceTypeStr = strings.ToUpper(serviceTypeStr)
		query = query.Where("service_type = ?", serviceTypeStr)
	}

	if err := query.Offset(offset).Limit(limit).Order("name ASC").Find(&tenders).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при получении тендеров.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	tenderResponses := make([]models.TenderResponse, len(tenders))
	for i, tender := range tenders {
		tenderResponses[i] = models.TenderResponse{
			ID:          tender.ID.String(),
			Name:        tender.Name,
			Description: tender.Description,
			Status:      tender.Status,
			ServiceType: tender.ServiceType,
			Version:     tender.Version,
			CreatedAt:   tender.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tenderResponses)
}

func GetUserTendersHandler(w http.ResponseWriter, r *http.Request) {

	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Параметр username обязателен.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	log.Printf("Ищем пользователя с username: %s", username)

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	offset := 0

	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			errorResponse := models.NewErrorResponse("Неверный формат параметра limit.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
	}

	if offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			w.WriteHeader(http.StatusBadRequest)
			errorResponse := models.NewErrorResponse("Неверный формат параметра offset.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
	}

	// 1. Найти пользователя по username
	var employee models.Employee
	if err := db.DB.Where("username = ?", username).First(&employee).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			errorResponse := models.NewErrorResponse("Пользователь не найден.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Ошибка при получении пользователя: %v", err)
		errorResponse := models.NewErrorResponse("Ошибка при получении пользователя.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// 2. Найти OrganizationID по UserID в OrganizationResponsible
	var orgResponsible models.OrganizationResponsible
	if err := db.DB.Where("user_id = ?", employee.ID).First(&orgResponsible).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			errorResponse := models.NewErrorResponse("Организация не найдена для данного пользователя.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Ошибка при получении организации: %v", err)
		errorResponse := models.NewErrorResponse("Ошибка при получении организации.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// 3. Найти все тендеры по OrganizationID
	var tenders []models.Tender
	if err := db.DB.Where("organization_id = ?", orgResponsible.OrganizationID).Offset(offset).Limit(limit).Order("name ASC").Find(&tenders).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			errorResponse := models.NewErrorResponse("Тендеры не найдены.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Ошибка при получении тендеров: %v", err)
		errorResponse := models.NewErrorResponse("Ошибка при получении тендеров.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	tenderResponses := make([]models.TenderResponse, len(tenders))
	for i, tender := range tenders {
		tenderResponses[i] = models.TenderResponse{
			ID:          tender.ID.String(),
			Name:        tender.Name,
			Description: tender.Description,
			Status:      tender.Status,
			ServiceType: tender.ServiceType,
			Version:     tender.Version,
			CreatedAt:   tender.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tenderResponses)
}

func GetTenderStatusHandler(w http.ResponseWriter, r *http.Request) {

	tenderIdStr := r.URL.Path[len("/api/tenders/") : len("/api/tenders/")+36]
	tenderId, err := uuid.Parse(tenderIdStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Неверный формат идентификатора тендера.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Параметр username обязателен.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	var tender models.Tender
	if err := db.DB.First(&tender, "id = ?", tenderId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			errorResponse := models.NewErrorResponse("Тендер не найден.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при получении статуса тендера.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tender.Status)
}

func UpdateTenderStatusHandler(w http.ResponseWriter, r *http.Request) {

	tenderIdStr := r.URL.Path[len("/api/tenders/") : len("/api/tenders/")+36]
	tenderId, err := uuid.Parse(tenderIdStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Неверный формат идентификатора тендера.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Параметр status обязателен.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	status = strings.ToUpper(status)

	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Параметр username обязателен.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	var tender models.Tender
	if err := db.DB.First(&tender, "id = ?", tenderId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			errorResponse := models.NewErrorResponse("Тендер не найден.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при получении тендера.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	tender.Status = models.TenderStatus(status)

	if err := db.DB.Save(&tender).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при обновлении статуса тендера.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	tenderResponse := models.TenderResponse{
		ID:          tender.ID.String(),
		Name:        tender.Name,
		Description: tender.Description,
		Status:      tender.Status,
		ServiceType: tender.ServiceType,
		Version:     tender.Version,
		CreatedAt:   tender.CreatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tenderResponse)
}

func EditTenderHandler(w http.ResponseWriter, r *http.Request) {

	tenderIdStr := r.URL.Path[len("/api/tenders/") : len("/api/tenders/")+36]
	tenderId, err := uuid.Parse(tenderIdStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Неверный формат идентификатора тендера.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Параметр username обязателен.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка чтения тела запроса.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	log.Printf("Полученные данные: %s", string(body))

	re := regexp.MustCompile(`[^\x00-\x7F]+`)
	cleanBody := re.ReplaceAllString(string(body), "")

	var updateData models.NewTenderRequest
	if err := json.Unmarshal([]byte(cleanBody), &updateData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Ошибка декодирования данных: %v", err)
		errorResponse := models.NewErrorResponse("Данные неправильно сформированы или не соответствуют требованиям.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	var tender models.Tender
	if err := db.DB.First(&tender, "id = ?", tenderId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			errorResponse := models.NewErrorResponse("Тендер не найден.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при получении тендера.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	tenderVersion := models.TenderVersion{
		TenderID:       tender.ID.String(),
		Name:           tender.Name,
		Description:    tender.Description,
		Status:         tender.Status,
		ServiceType:    tender.ServiceType,
		OrganizationID: tender.OrganizationID,
		Version:        tender.Version,
		CreatedAt:      tender.CreatedAt,
	}

	if err := db.DB.Create(&tenderVersion).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при сохранении версии тендера.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	newTender := models.Tender{
		ID:             uuid.New(),
		Name:           updateData.Name,
		Description:    updateData.Description,
		Status:         tender.Status,
		ServiceType:    updateData.ServiceType,
		OrganizationID: tender.OrganizationID,
		Version:        tender.Version + 1,
		CreatedAt:      time.Now(),
	}

	if err := db.DB.Create(&newTender).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при создании новой версии тендера.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	tenderResponse := models.TenderResponse{
		ID:          newTender.ID.String(),
		Name:        newTender.Name,
		Description: newTender.Description,
		Status:      newTender.Status,
		ServiceType: newTender.ServiceType,
		Version:     newTender.Version,
		CreatedAt:   newTender.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tenderResponse)
}

func RollbackTenderHandler(w http.ResponseWriter, r *http.Request) {

	tenderIdStr := r.URL.Path[len("/api/tenders/") : len("/api/tenders/")+36]
	tenderId, err := uuid.Parse(tenderIdStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Неверный формат идентификатора тендера.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	versionStr := r.URL.Path[len("/api/tenders/"+tenderIdStr+"/rollback/"):]
	version, err := strconv.Atoi(versionStr)
	if err != nil || version <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Неверный формат версии.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Параметр username обязателен.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	var previousTender models.TenderVersion
	if err := db.DB.Where("tender_id = ? AND version = ?", tenderId, version).First(&previousTender).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			errorResponse := models.NewErrorResponse("Предыдущая версия тендера не найдена.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при получении предыдущей версии тендера.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	newTender := models.Tender{
		ID:             uuid.New(),
		Name:           previousTender.Name,
		Description:    previousTender.Description,
		Status:         previousTender.Status,
		ServiceType:    previousTender.ServiceType,
		OrganizationID: previousTender.OrganizationID,
		Version:        previousTender.Version,
		CreatedAt:      previousTender.CreatedAt,
	}

	if err := db.DB.Create(&newTender).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при создании новой версии тендера.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	tenderResponse := models.TenderResponse{
		ID:          newTender.ID.String(),
		Name:        newTender.Name,
		Description: newTender.Description,
		Status:      newTender.Status,
		ServiceType: newTender.ServiceType,
		Version:     newTender.Version,
		CreatedAt:   newTender.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tenderResponse)
}

func CreateBidHandler(w http.ResponseWriter, r *http.Request) {
	var newBidRequest models.NewBidRequest
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.NewErrorResponse("Ошибка чтения тела запроса."))
		return
	}

	re := regexp.MustCompile(`[^\x00-\x7F]+`)
	cleanBody := re.ReplaceAllString(string(body), "")

	if err := json.Unmarshal([]byte(cleanBody), &newBidRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.NewErrorResponse("Ошибка декодирования JSON: " + err.Error()))
		return
	}

	if newBidRequest.Name == "" || newBidRequest.Description == "" || newBidRequest.TenderID == "" || newBidRequest.AuthorType == "" || newBidRequest.AuthorID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.NewErrorResponse("Поля name, description, tenderId, authorType и authorId обязательны."))
		return
	}
	var tender models.Tender
	var bid models.Bid
	if err := db.DB.First(&tender, "id = ?", newBidRequest.TenderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			errorResponse := models.NewErrorResponse("Тендер не найден.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при проверке существования тендера.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	newbid := models.Bid{
		ID:          uuid.New(),
		Name:        newBidRequest.Name,
		Description: newBidRequest.Description,
		Status:      models.BID_CREATED,
		TenderID:    newBidRequest.TenderID,
		AuthorType:  newBidRequest.AuthorType,
		AuthorID:    newBidRequest.AuthorID,
		Version:     bid.Version,
		CreatedAt:   time.Now(),
	}

	if err := db.DB.Create(&newbid).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.NewErrorResponse("Сервер не готов обрабатывать запросы."))
		return
	}

	bidRespone := models.Bid{
		ID:         bid.ID,
		Name:       newBidRequest.Name,
		Status:     models.BID_CREATED,
		AuthorType: newBidRequest.AuthorType,
		AuthorID:   newBidRequest.AuthorID,
		Version:    bid.Version,
		CreatedAt:  bid.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bidRespone)
}

func GetUserBidsHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры запроса
	username := r.URL.Query().Get("username")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Параметр username обязателен.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Парсим limit и offset
	limit := 10 // Значение по умолчанию
	offset := 0

	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			w.WriteHeader(http.StatusBadRequest)
			errorResponse := models.NewErrorResponse("Некорректное значение параметра limit.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
	}

	if offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			w.WriteHeader(http.StatusBadRequest)
			errorResponse := models.NewErrorResponse("Некорректное значение параметра offset.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
	}

	log.Printf("Ищем заявки для пользователя с username: %s", username)

	// Шаг 1: Получаем UserID из таблицы Employee по username
	var employee models.Employee
	if err := db.DB.Where("username = ?", username).First(&employee).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			errorResponse := models.NewErrorResponse("Пользователь не найден.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при получении пользователя.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Шаг 2: Получаем все bids по AuthorID (который равен UserID)
	var bids []models.Bid
	if err := db.DB.Where("author_id = ?", employee.ID).
		Limit(limit).
		Offset(offset).
		Find(&bids).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			errorResponse := models.NewErrorResponse("Заявки не найдены.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при получении заявок.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Формируем ответ
	bidResponses := make([]models.BidResponse, len(bids))
	for i, bid := range bids {
		bidResponses[i] = models.BidResponse{
			ID:         bid.ID.String(),
			Name:       bid.Name,
			Status:     bid.Status,
			AuthorType: bid.AuthorType,
			AuthorID:   bid.AuthorID,
			Version:    bid.Version,
			CreatedAt:  bid.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bidResponses)
}

func GetBidsForTenderHandler(w http.ResponseWriter, r *http.Request) {

	tenderIdStr := r.URL.Path[len("/api/bids/") : len("/api/bids/")+36]
	tenderId, err := uuid.Parse(tenderIdStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Неверный формат идентификатора тендера.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Параметр username обязателен.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	offset := 0

	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			errorResponse := models.NewErrorResponse("Неверный формат параметра limit.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
	}

	if offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			w.WriteHeader(http.StatusBadRequest)
			errorResponse := models.NewErrorResponse("Неверный формат параметра offset.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
	}

	log.Printf("Ищем заявки для тендера с ID: %s и пользователя с username: %s", tenderId.String(), username)

	// Шаг 1: Получаем UserID из таблицы Employee по username
	var employee models.Employee
	if err := db.DB.Where("username = ?", username).First(&employee).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			errorResponse := models.NewErrorResponse("Пользователь не найден.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при получении пользователя.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Шаг 2: Получаем все bids по AuthorID (который равен UserID) и TenderID
	var bids []models.Bid
	if err := db.DB.Where("author_id = ? AND tender_id = ?", employee.ID, tenderId).
		Limit(limit).
		Offset(offset).
		Order("name ASC").
		Find(&bids).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			errorResponse := models.NewErrorResponse("Заявки не найдены.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при получении заявок.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Формируем ответ
	bidResponses := make([]models.BidResponse, len(bids))
	for i, bid := range bids {
		bidResponses[i] = models.BidResponse{
			ID:         bid.ID.String(),
			Name:       bid.Name,
			Status:     bid.Status,
			AuthorType: bid.AuthorType,
			AuthorID:   bid.AuthorID,
			Version:    bid.Version,
			CreatedAt:  bid.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bidResponses)
}

func GetBidStatusHandler(w http.ResponseWriter, r *http.Request) {

	bidIdStr := r.URL.Path[len("/api/bids/") : len("/api/bids/")+36]
	bidId, err := uuid.Parse(bidIdStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Неверный формат идентификатора предложения.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Параметр username обязателен.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	var bid models.Bid
	if err := db.DB.First(&bid, "id = ?", bidId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			errorResponse := models.NewErrorResponse("Предложение не найдено.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при получении статуса предложения.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bid.Status)
}

func UpdateBidStatusHandler(w http.ResponseWriter, r *http.Request) {

	bidIdStr := r.URL.Path[len("/api/bids/") : len("/api/bids/")+36]
	bidId, err := uuid.Parse(bidIdStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Неверный формат идентификатора предложения.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Параметр status обязателен.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	status = strings.ToUpper(status)

	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Параметр username обязателен.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	var bid models.Bid
	if err := db.DB.First(&bid, "id = ?", bidId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			errorResponse := models.NewErrorResponse("Предложение не найдено.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при получении предложения.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	bid.Status = models.BidStatus(status)

	if err := db.DB.Save(&bid).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при обновлении статуса предложения.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	bidResponses := models.BidResponse{
		ID:         bid.ID.String(),
		Name:       bid.Name,
		Status:     bid.Status,
		AuthorType: bid.AuthorType,
		AuthorID:   bid.AuthorID,
		Version:    bid.Version,
		CreatedAt:  bid.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bidResponses)
}

func EditBidHandler(w http.ResponseWriter, r *http.Request) {

	bidIdStr := r.URL.Path[len("/api/bids/") : len("/api/bids/")+36]
	bidId, err := uuid.Parse(bidIdStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Неверный формат идентификатора предложения.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Параметр username обязателен.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка чтения тела запроса.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	log.Printf("Полученные данные: %s", string(body))

	re := regexp.MustCompile(`[^\x00-\x7F]+`)
	cleanBody := re.ReplaceAllString(string(body), "")

	var updateData models.NewBidRequest
	if err := json.Unmarshal([]byte(cleanBody), &updateData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Ошибка декодирования данных: %v", err)
		errorResponse := models.NewErrorResponse("Данные неправильно сформированы или не соответствуют требованиям.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	var bid models.Bid
	if err := db.DB.First(&bid, "id = ?", bidId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			errorResponse := models.NewErrorResponse("Предложение не найдено.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при получении предложения.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	bidVersion := models.BidVersion{
		BidID:       bid.ID.String(),
		Name:        bid.Name,
		Description: bid.Description,
		Status:      bid.Status,
		TenderID:    bid.TenderID,
		AuthorType:  bid.AuthorType,
		AuthorID:    bid.AuthorID,
		Version:     bid.Version,
		CreatedAt:   bid.CreatedAt,
	}

	if err := db.DB.Create(&bidVersion).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при сохранении версии предложения.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	newBid := models.Bid{
		ID:          uuid.New(),
		Name:        updateData.Name,
		Description: updateData.Description,
		Status:      bid.Status,
		TenderID:    bid.TenderID,
		AuthorType:  bid.AuthorType,
		AuthorID:    bid.AuthorID,
		Version:     bid.Version + 1,
		CreatedAt:   time.Now(),
	}

	if err := db.DB.Create(&newBid).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при создании новой версии предложения.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	bidResponses := models.BidResponse{
		ID:         newBid.ID.String(),
		Name:       newBid.Name,
		Status:     newBid.Status,
		AuthorType: newBid.AuthorType,
		AuthorID:   newBid.AuthorID,
		Version:    newBid.Version,
		CreatedAt:  newBid.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bidResponses)
}

func RollbackBidHandler(w http.ResponseWriter, r *http.Request) {

	bidIdStr := r.URL.Path[len("/api/bids/") : len("/api/bids/")+36]
	bidId, err := uuid.Parse(bidIdStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Неверный формат идентификатора предложения.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	versionStr := r.URL.Path[len("/api/bids/"+bidIdStr+"/rollback/"):]
	version, err := strconv.Atoi(versionStr)
	if err != nil || version <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Неверный формат версии.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := models.NewErrorResponse("Параметр username обязателен.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	var previousBid models.BidVersion
	if err := db.DB.Where("bid_id = ? AND version = ?", bidId, version).First(&previousBid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			errorResponse := models.NewErrorResponse("Предыдущая версия предложения не найдена.")
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при получении предыдущей версии предложения.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	newBid := models.Bid{
		ID:          uuid.New(),
		Name:        previousBid.Name,
		Description: previousBid.Description,
		Status:      previousBid.Status,
		TenderID:    previousBid.TenderID,
		AuthorType:  previousBid.AuthorType,
		AuthorID:    previousBid.AuthorID,
		Version:     previousBid.Version,
		CreatedAt:   previousBid.CreatedAt,
	}

	if err := db.DB.Create(&newBid).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := models.NewErrorResponse("Ошибка при создании новой версии тендера.")
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	bidResponses := models.BidResponse{
		ID:         newBid.ID.String(),
		Name:       newBid.Name,
		Status:     newBid.Status,
		AuthorType: newBid.AuthorType,
		AuthorID:   newBid.AuthorID,
		Version:    newBid.Version,
		CreatedAt:  newBid.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bidResponses)
}
