package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"shipment.com/packages/distance"
	"shipment.com/packages/location"
	"shipment.com/packages/models"
)

// Trip represents the structure of a trip record
type Trip struct {
	TripNumber   string       `bson:"tripnumber"`
	IndentNumber int          `bson:"indentnumber"`
	DemandZone   string       `bson:"demandzone"`
	Placesdata   []Place      `bson:"placesdata"`
	ShipperData  ShipperData  `bson:"shipperdata"`
	PricesData   PricesData   `bson:"pricesdata"`
	SupplierData SupplierData `bson:"supplierdata"`
	VehicleData  VehicleData  `bson:"vehicledata"`
	CreatedAt    time.Time    `json:"createdat"`
}
type Triplog struct {
	Tripnumber       string    `bson:"tripnumber"`
	AssigneeMatching int       `json:"assigneematching"`
	AssigneeDemand   int       `json:"assigneedemand"`
	CreatedAt        time.Time `json:"createdat"`
}
type Place struct {
	Origin      Origin      `bson:"origin"`
	Destination Destination `bson:"destination"`
	Distance    int         `bson:"distance"`
}

type SupplierData struct {
	UUID int    `json:"supplieruuid"`
	Name string `json:"name"`
}
type VehicleData struct {
	Type   string `json:"vt"`
	Number string `json:"vehiclenumber"`
}

type Origin struct {
	UUID  int64  `bson:"uuid"`
	Label string `bson:"label"`
}

type Destination struct {
	UUID  int    `bson:"uuid"`
	Label string `bson:"label"`
}
type ShipperData struct {
	ShipperUUID int    `bson:"shipperuuid"`
	BranchUUID  int    `bson:"branchuuid"`
	ShipperName string `bson:"shippername"`
	BranchName  string `bson:"branchname"`
}

type PricesData struct {
	FreightPrice     float64 `bson:"freightprice"`
	TransportPrice   float64 `bson:"transportprice"`
	MarginValidation float64 `bson:"marginvalidation"`
}

var (
	counter int64
	mutex   sync.Mutex
)

// generateTripID generates a unique trip ID
/*func generateTripID() string {
	mutex.Lock()
	defer mutex.Unlock()
	counter++
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("TRIP-%s-%06d", timestamp, counter)
}*/

func getNextSequence() string {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = mongo.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	collection := client.Database("Raaho").Collection("countertriplog")

	update := bson.M{"$inc": bson.M{"seq": 1}}
	opt := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)

	var counter models.Counter
	err = collection.FindOneAndUpdate(context.Background(), bson.M{"id": "trip"}, update, opt).Decode(&counter)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("TRIP-%d", counter.Seq)
}
func main() {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = mongo.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	// Get collections
	placementCollection := client.Database("Raaho").Collection("placements")
	tripCollection := client.Database("Raaho").Collection("trip")
	triplogCollection := client.Database("Raaho").Collection("triplog")

	// Filter placements with status 0
	filter := bson.M{"status": 0}

	// Get cursor for unprocessed placements
	cursor, err := placementCollection.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	// Iterate through unprocessed placements
	for cursor.Next(ctx) {
		var placement models.Placement
		err := cursor.Decode(&placement)
		if err != nil {
			log.Fatal(err)
		}

		trip := createTrip(placement)
		triplog := createTripLog(trip, placement)

		insertResult, err := tripCollection.InsertOne(ctx, trip)
		if err != nil {
			log.Fatal(err)
		}
		_, err = triplogCollection.InsertOne(ctx, triplog)

		fmt.Println("Inserted log record with ID:", insertResult.InsertedID)

		// Update placement status to 1 after creating the trip log
		update := bson.M{"$set": bson.M{"status": 1}}
		_, err = placementCollection.UpdateOne(ctx, bson.M{"placementid": placement.PlacementID}, update)
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}
}
func createTripLog(trip Trip, shipment models.Placement) Triplog {
	return Triplog{
		Tripnumber:       trip.TripNumber,
		AssigneeMatching: shipment.AssigneeMatching,
		AssigneeDemand:   shipment.AssigneeDemand,
		CreatedAt:        time.Now(),
	}

}

func createTrip(placement models.Placement) Trip {
	return Trip{
		TripNumber:   getNextSequence(),
		IndentNumber: placement.IndentNumber,
		DemandZone:   placement.DemandZone,
		Placesdata: []Place{
			{
				Origin: Origin{
					UUID:  placement.OriginUUID,
					Label: location.GetLabelByID(int(placement.OriginUUID)),
				},
				Destination: Destination{
					UUID:  placement.DestinationUUID,
					Label: location.GetLabelByID(int(placement.DestinationUUID)),
				},
				Distance: distance.GetDistance(location.GetLabelByID(int(placement.OriginUUID)), location.GetLabelByID(int(placement.DestinationUUID))),
			},
		},
		ShipperData: ShipperData{
			ShipperUUID: placement.SupplierUUID,
			BranchUUID:  54321,
			ShipperName: "ACME Logistics",
			BranchName:  "North Branch",
		},
		PricesData: PricesData{
			FreightPrice:     placement.FreightPrice,
			TransportPrice:   placement.TransportPrice,
			MarginValidation: placement.TransportPrice - placement.FreightPrice,
		},
		SupplierData: SupplierData{
			UUID: placement.SupplierUUID,
			Name: "AHAD",
		},
		VehicleData: VehicleData{
			Type:   placement.VehicleType,
			Number: placement.VehicleNumber,
		},
		CreatedAt: time.Now(),
	}
}
