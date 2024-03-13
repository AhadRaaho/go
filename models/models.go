package models

type Placement struct {
	PlacementID      int     `json:"placementId"`
	OriginUUID       int64   `json:"originUuid"`
	DestinationUUID  int     `json:"destinationUuid"`
	CustomerBranch   int     `json:"customerBranch"`
	SupplierUUID     int     `json:"supplierUuid"`
	FreightPrice     float64 `json:"freightPrice"`
	TransportPrice   float64 `json:"transportPrice"`
	DemandZone       string  `json:"demandZone"`
	IndentNumber     int     `json:"indentNumber"`
	AssigneeMatching int     `json:"assigneeMatching"`
	AssigneeDemand   int     `json:"assigneeDemand"`
	VehicleType      string  `json:"vehicleType"`
	VehicleNumber    string  `json:"vehicleNumber"`
	Status           int     `json:"status" default:"0"`
}

type Counter struct {
	ID  string `bson:"id,omitempty"`
	Seq int    `bson:"seq"`
}

type Place struct {
	PlaceId           string  `json:"placeId"`
	PlaceName         string  `json:"placeName"`
	PlaceLatitude     float64 `json:"placeLatitude"`
	PlaceLongitude    float64 `json:"placeLongitude"`
	PrimaryZoneCode   string  `json:"primaryZoneCode"`
	PrimaryZoneName   string  `json:"primaryZoneName"`
	SecondaryZoneCode string  `json:"secondaryZoneCode"`
	SecondaryZoneName string  `json:"secondaryZoneName"`
	RegionCode        string  `json:"regionCode"`
	RegionName        string  `json:"regionName"`
	InternalZoneCode  string  `json:"internalZoneCode"`
	InternalZoneName  string  `json:"internalZoneName"`
	InternalZoneState string  `json:"internalZoneState"`
	GooglePlaceId     string  `json:"googlePlaceId"`
	LinkedAddress     string  `json:"linkedAddress"`
	S2Token           string  `json:"s2Token"`
}

type ApiResponse struct {
	Status    string   `json:"status"`
	Error     string   `json:"error"`
	Data      []Place  `json:"data"`
	ErrorList []string `json:"errorList"`
}
