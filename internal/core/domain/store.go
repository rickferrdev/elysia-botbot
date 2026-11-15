package domain

import "time"

type Inventories struct {
	ID        int
	UserID    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type InventoriesItems struct {
	ID          int
	InventoryID int
	Quantity    int
	ItemID      int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CatalogType string

type Catalog struct {
	ID        int
	Name      string
	Type      CatalogType
	Stackable int
}

type StoreListings struct {
	ID                int
	ItemID            int
	QuantityAvailable int
	Enabled           bool
}

type StoreListingsCost struct {
	ID        int
	ListingID int
	ItemID    int
	Quantity  int
}
