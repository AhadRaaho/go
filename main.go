package main

import (
	"context"
	"fmt"

	//"database"
	//"fmt"

	"shipment.com/packages/database"

	"github.com/gin-gonic/gin"
	"shipment.com/packages/placementfunctions"
)

func main() {
	r := gin.Default()
	r1 := database.Sum()
	fmt.Println(r1)

	placementsDB, err := database.ConnectToMongoDB()
	if err != nil {
		panic(err)
	}
	defer placementsDB.Disconnect(context.Background())

	placementcollection := placementsDB.Database("Raaho").Collection("placements")
	r.POST("/placements", func(c *gin.Context) { placementfunctions.CreatePlacement(c, placementsDB, placementcollection) })
	r.GET("/placements", func(c *gin.Context) { placementfunctions.GetPlacement(c, placementcollection) })
	r.GET("/placements/:placementid", func(c *gin.Context) { placementfunctions.GetPlacementByPlacementID(c, placementcollection) })
	r.GET("/placements/search/origin/:originuuid", func(c *gin.Context) { placementfunctions.GetPlacementsByOriginUUID(c, placementcollection) })
	r.GET("/placements/search/destination/:destinationuuid", func(c *gin.Context) { placementfunctions.GetPlacementsByDestinationUUID(c, placementcollection) })
	r.GET("/placements/search/customerbranch/:customerbranch", func(c *gin.Context) { placementfunctions.GetPlacementsByCustomerBranch(c, placementcollection) })
	r.GET("/placements/search/demandzone/:demandzone", func(c *gin.Context) { placementfunctions.GetPlacementsByDemandZone(c, placementcollection) })
	r.PATCH("/placements/:placementid", func(c *gin.Context) { placementfunctions.PatchPlacement(c, placementsDB, placementcollection) })
	r.PUT("/placements/:placementid", func(c *gin.Context) { placementfunctions.Updatedplacement(c, placementsDB, placementcollection) })

	r.Run("localhost:9090")
}
