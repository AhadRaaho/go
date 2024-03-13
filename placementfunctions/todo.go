package placementfunctions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"shipment.com/packages/models"
)

func getNextSequence(client *mongo.Client, collectionName string) (int, error) {
	collection := client.Database("Raaho").Collection("counters")

	update := bson.M{"$inc": bson.M{"seq": 1}}
	opt := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)

	var counter models.Counter
	err := collection.FindOneAndUpdate(context.Background(), bson.M{"id": collectionName}, update, opt).Decode(&counter)
	if err != nil {
		return 0, err
	}

	return counter.Seq, nil
}
func GetZoneByID(id int) string {
	placeId := strconv.Itoa(id)
	url := fmt.Sprintf("https://api.stage.raaho.in/places2/v2/places/%s", placeId)

	response, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return ""
	}

	var apiResponse models.ApiResponse
	err = json.NewDecoder(response.Body).Decode(&apiResponse)
	if err != nil {
		return ""
	}

	if len(apiResponse.Data) == 0 {
		return ""
	}

	return apiResponse.Data[0].PrimaryZoneName
}
func CreatePlacement(c *gin.Context, client *mongo.Client, collection *mongo.Collection) {
	var placement models.Placement
	err := c.BindJSON(&placement)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nextID, err := getNextSequence(client, "placementId")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate ID"})
		return
	}

	placement.PlacementID = nextID
	//fmt.Print(placement)
	demandZone := GetZoneByID(int(placement.OriginUUID))
	//placement.DemandZone = GetZoneByID(int(placement.OriginUUID))
	if demandZone == "" {
		// Handle error from GetZoneByID if needed
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get demand zone"})
		return
	}
	placement.DemandZone = demandZone

	_, err = collection.InsertOne(context.Background(), placement)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "placement created successfully", "data": placement})
}

func GetPlacement(c *gin.Context, collection *mongo.Collection) {
	// Parse query parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}

	perPage, err := strconv.Atoi(c.DefaultQuery("perPage", "10"))
	if err != nil || perPage < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid perPage value"})
		return
	}

	// Calculate the offset based on page and perPage
	offset := (page - 1) * perPage

	cursor, err := collection.Find(context.Background(), bson.M{}, options.Find().SetLimit(int64(perPage)).SetSkip(int64(offset)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var placements []models.Placement
	err = cursor.All(context.Background(), &placements)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get total count of items
	totalItems, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate total pages
	totalPages := (totalItems + int64(perPage) - 1) / int64(perPage)

	c.JSON(http.StatusOK, gin.H{
		"message": "Data retrieved successfully",
		"data":    placements,
		"pagination": gin.H{
			"page":       page,
			"perPage":    perPage,
			"totalPages": totalPages,
			"totalItems": totalItems,
		},
	})
}

func Updatedplacement(c *gin.Context, client *mongo.Client, collection *mongo.Collection) {
	id := c.Param("placementid")

	placementID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Placement ID"})
		return
	}
	var updatedplacement models.Placement
	if err := c.ShouldBindJSON(&updatedplacement); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := bson.M{"placementid": placementID}
	update := bson.M{"$set": bson.M{
		"originuuid":       updatedplacement.OriginUUID,
		"destinationuuid":  updatedplacement.DestinationUUID,
		"customerbranch":   updatedplacement.CustomerBranch,
		"supplieruuid":     updatedplacement.SupplierUUID,
		"freightprice":     updatedplacement.FreightPrice,
		"transportprice":   updatedplacement.TransportPrice,
		"demandzone":       updatedplacement.DemandZone,
		"indentnumber":     updatedplacement.IndentNumber,
		"assigneematching": updatedplacement.AssigneeMatching,
		"assigneedemand":   updatedplacement.AssigneeDemand,
		"vehicletype":      updatedplacement.VehicleType,
		"vehiclenumber":    updatedplacement.VehicleNumber,
		"status":           updatedplacement.Status,
	}}

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.ModifiedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Placement not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Placement updated successfully", "data": updatedplacement})
}

func GetPlacementByPlacementID(c *gin.Context, collection *mongo.Collection) {
	ID := c.Param("placementid")

	placementID, err := strconv.Atoi(ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid placement ID"})
		return
	}

	filter := bson.M{"placementid": placementID}

	var placement models.Placement
	err = collection.FindOne(context.Background(), filter).Decode(&placement)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "placement not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Placement retrieved successfully", "data": placement})
}

func GetPlacementsByOriginUUID(c *gin.Context, collection *mongo.Collection) {
	// Parse query parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}

	perPage, err := strconv.Atoi(c.DefaultQuery("perPage", "10"))
	if err != nil || perPage < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid perPage value"})
		return
	}

	// Calculate the offset based on page and perPage
	offset := (page - 1) * perPage

	originUUID, err := strconv.Atoi(c.Param("originuuid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid originuuid"})
		return
	}

	filter := bson.M{"originuuid": originUUID}

	cursor, err := collection.Find(context.Background(), filter, options.Find().SetLimit(int64(perPage)).SetSkip(int64(offset)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var placements []models.Placement
	err = cursor.All(context.Background(), &placements)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get total count of items
	totalItems, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate total pages
	totalPages := (totalItems + int64(perPage) - 1) / int64(perPage)

	c.JSON(http.StatusOK, gin.H{
		"message": "Placements retrieved successfully",
		"data":    placements,
		"pagination": gin.H{
			"page":       page,
			"perPage":    perPage,
			"totalPages": totalPages,
			"totalItems": totalItems,
		},
	})
}

func GetPlacementsByDestinationUUID(c *gin.Context, collection *mongo.Collection) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}

	perPage, err := strconv.Atoi(c.DefaultQuery("perPage", "10"))
	if err != nil || perPage < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid perPage value"})
		return
	}

	offset := (page - 1) * perPage

	destinationUUID, err := strconv.Atoi(c.Param("destinationuuid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid destinationuuid"})
		return
	}

	filter := bson.M{"destinationuuid": destinationUUID}

	cursor, err := collection.Find(context.Background(), filter, options.Find().SetLimit(int64(perPage)).SetSkip(int64(offset)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var placements []models.Placement
	err = cursor.All(context.Background(), &placements)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalItems, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (totalItems + int64(perPage) - 1) / int64(perPage)

	c.JSON(http.StatusOK, gin.H{
		"message": "Placements retrieved successfully",
		"data":    placements,
		"pagination": gin.H{
			"page":       page,
			"perPage":    perPage,
			"totalPages": totalPages,
			"totalItems": totalItems,
		},
	})
}

func GetPlacementsByCustomerBranch(c *gin.Context, collection *mongo.Collection) {

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}

	perPage, err := strconv.Atoi(c.DefaultQuery("perPage", "10"))
	if err != nil || perPage < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid perPage value"})
		return
	}

	offset := (page - 1) * perPage

	customerBranch, err := strconv.Atoi(c.Param("customerbranch"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Customer Branch"})
		return
	}

	filter := bson.M{"customerbranch": customerBranch}

	cursor, err := collection.Find(context.Background(), filter, options.Find().SetLimit(int64(perPage)).SetSkip(int64(offset)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var placements []models.Placement
	err = cursor.All(context.Background(), &placements)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalItems, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (totalItems + int64(perPage) - 1) / int64(perPage)

	c.JSON(http.StatusOK, gin.H{
		"message": "Placements retrieved successfully",
		"data":    placements,
		"pagination": gin.H{
			"page":       page,
			"perPage":    perPage,
			"totalPages": totalPages,
			"totalItems": totalItems,
		},
	})
}

func GetPlacementsByDemandZone(c *gin.Context, collection *mongo.Collection) {
	demand_Zone := c.Param("demandzone")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}

	perPage, err := strconv.Atoi(c.DefaultQuery("perPage", "10"))
	if err != nil || perPage < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid perPage value"})
		return
	}

	offset := (page - 1) * perPage

	filter := bson.M{"demandzone": demand_Zone}

	options := options.Find()
	options.SetLimit(int64(perPage))
	options.SetSkip(int64(offset))

	cursor, err := collection.Find(context.Background(), filter, options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var placements []models.Placement
	err = cursor.All(context.Background(), &placements)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalItems, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (totalItems + int64(perPage) - 1) / int64(perPage)

	c.JSON(http.StatusOK, gin.H{
		"message": "Placements retrieved successfully",
		"data":    placements,
		"pagination": gin.H{
			"page":       page,
			"perPage":    perPage,
			"totalPages": totalPages,
			"totalItems": totalItems,
		},
	})
}

func PatchPlacement(c *gin.Context, client *mongo.Client, collection *mongo.Collection) {
	placementID := c.Param("placementid")

	userID, err := strconv.Atoi(placementID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid placement ID"})
		return
	}

	var updatedFields bson.M
	if err := c.BindJSON(&updatedFields); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := bson.M{"placementid": userID}
	update := bson.M{"$set": updatedFields}

	// Set the options to return the updated document
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedPlacement models.Placement
	err = collection.FindOneAndUpdate(context.Background(), filter, update, options).Decode(&updatedPlacement)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Placement updated successfully", "data": updatedPlacement})
}
