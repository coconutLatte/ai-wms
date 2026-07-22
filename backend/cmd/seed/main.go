// Seed Script — populates the database with demo data for UI development and testing.
// Idempotent: skips seeding if demo data already exists.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository/postgres"
	"github.com/ai-wms/ai-wms/backend/internal/service"
	"github.com/ai-wms/ai-wms/backend/pkg/config"
	"github.com/ai-wms/ai-wms/backend/pkg/logger"
	"github.com/google/uuid"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "invalid configuration: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.LogLevel)

	// Connect to database.
	db, err := postgres.NewDB(context.Background(), cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Run database migrations (idempotent — only unapplied migrations execute).
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}
	if err := db.RunMigrationsFromDir(context.Background(), migrationsDir); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	warehouseRepo := postgres.NewWarehouseRepo(db)
	inventoryRepo := postgres.NewInventoryRepo(db)

	ctx := context.Background()

	// Always ensure admin user and default roles exist (before idempotency check).
	seedAdminUser(ctx, db, log)

	// ── Idempotency Check ──────────────────────────────────────
	warehouseCount, err := warehouseRepo.CountWarehouses(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to count warehouses: %v\n", err)
		os.Exit(1)
	}
	if warehouseCount > 0 {
		log.Info(fmt.Sprintf("Database already has %d warehouse(s) — seed skipped", warehouseCount))
		os.Exit(0)
	}

	log.Info("Seeding demo data...")

	// ── 1. Demo Warehouse ──────────────────────────────────────
	warehouse := seedWarehouse(ctx, warehouseRepo)
	log.Info(fmt.Sprintf("  ✓ Warehouse: %s (%s)", warehouse.Name, warehouse.Code))

	// ── 2. Zones ───────────────────────────────────────────────
	zones := seedZones(ctx, warehouseRepo, warehouse.ID, log)

	// ── 3. Locations ───────────────────────────────────────────
	locationMap := seedLocations(ctx, warehouseRepo, warehouse.ID, zones, log)

	// ── 4. SKUs ────────────────────────────────────────────────
	skus := seedSKUs(ctx, inventoryRepo, log)

	// ── 5. Inventory ───────────────────────────────────────────
	seedInventory(ctx, inventoryRepo, warehouse.ID, skus, locationMap, log)

	// ── Summary ────────────────────────────────────────────────
	log.Info("")
	log.Info("Seed data loaded successfully!")
	log.Info(fmt.Sprintf("  Warehouse: %d", 1))
	log.Info(fmt.Sprintf("  Zones:     %d", len(zones)))
	log.Info(fmt.Sprintf("  Locations: %d", totalLocations()))
	log.Info(fmt.Sprintf("  SKUs:      %d", len(skus)))
	log.Info("  Admin user password set to: admin123")
	log.Info("")
	log.Info("Run 'make run-admin' to start the admin server.")
}

// ── Seeders ────────────────────────────────────────────────────────────────────

func seedWarehouse(ctx context.Context, repo *postgres.WarehouseRepo) *domain.Warehouse {
	w := &domain.Warehouse{
		Code:    "WH-MAIN",
		Name:    "Main Distribution Center",
		Address: "No. 88 Logistics Road, Shanghai Pilot Free Trade Zone",
		Status:  domain.WarehouseStatusActive,
	}
	if err := repo.CreateWarehouse(ctx, w); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create warehouse: %v\n", err)
		os.Exit(1)
	}
	return w
}

func seedZones(ctx context.Context, repo *postgres.WarehouseRepo, warehouseID uuid.UUID, log *logger.Logger) []*domain.Zone {
	zoneDefs := []struct {
		Code     string
		Name     string
		ZoneType domain.ZoneType
	}{
		{"Z-RCV", "Receiving Zone", domain.ZoneTypeReceiving},
		{"Z-STG", "Staging Zone", domain.ZoneTypeStaging},
		{"Z-STO-A", "Storage Zone A (Ambient)", domain.ZoneTypeStorage},
		{"Z-STO-B", "Storage Zone B (Ambient)", domain.ZoneTypeStorage},
		{"Z-PICK", "Picking Zone", domain.ZoneTypePicking},
		{"Z-SHIP", "Shipping Zone", domain.ZoneTypeShipping},
		{"Z-RET", "Returns Zone", domain.ZoneTypeReturns},
	}

	var zones []*domain.Zone
	for _, zd := range zoneDefs {
		z := &domain.Zone{
			WarehouseID: warehouseID,
			Code:        zd.Code,
			Name:        zd.Name,
			ZoneType:    zd.ZoneType,
			Status:      domain.ZoneStatusActive,
		}
		if err := repo.CreateZone(ctx, z); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create zone %s: %v\n", zd.Code, err)
			os.Exit(1)
		}
		zones = append(zones, z)
		log.Info(fmt.Sprintf("  ✓ Zone: %s (%s)", z.Name, z.Code))
	}
	return zones
}

func seedLocations(ctx context.Context, repo *postgres.WarehouseRepo, warehouseID uuid.UUID, zones []*domain.Zone, log *logger.Logger) map[string]*domain.Location {
	type locDef struct {
		Code         string
		ZoneCode     string
		LocationType domain.LocationType
		Capacity     *domain.Capacity
	}

	defs := []locDef{
		// Receiving docks
		{"RCV-DOCK-01", "Z-RCV", domain.LocationTypeFloor, nil},
		{"RCV-DOCK-02", "Z-RCV", domain.LocationTypeFloor, nil},
		{"RCV-DOCK-03", "Z-RCV", domain.LocationTypeFloor, nil},

		// Staging areas
		{"STG-01", "Z-STG", domain.LocationTypeFloor, nil},
		{"STG-02", "Z-STG", domain.LocationTypeFloor, nil},
		{"STG-03", "Z-STG", domain.LocationTypeFloor, nil},

		// Storage Zone A — pallet racks
		{"A-01-01-01", "Z-STO-A", domain.LocationTypePallet, &domain.Capacity{MaxWeight: 1000, MaxVolume: 2.0, MaxQty: 72}},
		{"A-01-01-02", "Z-STO-A", domain.LocationTypePallet, &domain.Capacity{MaxWeight: 1000, MaxVolume: 2.0, MaxQty: 72}},
		{"A-01-02-01", "Z-STO-A", domain.LocationTypePallet, &domain.Capacity{MaxWeight: 1000, MaxVolume: 2.0, MaxQty: 72}},
		{"A-01-02-02", "Z-STO-A", domain.LocationTypePallet, &domain.Capacity{MaxWeight: 1000, MaxVolume: 2.0, MaxQty: 72}},
		{"A-01-03-01", "Z-STO-A", domain.LocationTypePallet, &domain.Capacity{MaxWeight: 1000, MaxVolume: 2.0, MaxQty: 72}},
		{"A-01-03-02", "Z-STO-A", domain.LocationTypePallet, &domain.Capacity{MaxWeight: 1000, MaxVolume: 2.0, MaxQty: 72}},
		{"A-02-01-01", "Z-STO-A", domain.LocationTypeShelf, &domain.Capacity{MaxWeight: 500, MaxVolume: 1.0, MaxQty: 200}},
		{"A-02-01-02", "Z-STO-A", domain.LocationTypeShelf, &domain.Capacity{MaxWeight: 500, MaxVolume: 1.0, MaxQty: 200}},
		{"A-02-01-03", "Z-STO-A", domain.LocationTypeShelf, &domain.Capacity{MaxWeight: 500, MaxVolume: 1.0, MaxQty: 200}},
		{"A-02-01-04", "Z-STO-A", domain.LocationTypeShelf, &domain.Capacity{MaxWeight: 500, MaxVolume: 1.0, MaxQty: 200}},

		// Storage Zone B — pallet racks
		{"B-01-01-01", "Z-STO-B", domain.LocationTypePallet, &domain.Capacity{MaxWeight: 1000, MaxVolume: 2.0, MaxQty: 72}},
		{"B-01-01-02", "Z-STO-B", domain.LocationTypePallet, &domain.Capacity{MaxWeight: 1000, MaxVolume: 2.0, MaxQty: 72}},
		{"B-01-02-01", "Z-STO-B", domain.LocationTypePallet, &domain.Capacity{MaxWeight: 1000, MaxVolume: 2.0, MaxQty: 72}},
		{"B-01-02-02", "Z-STO-B", domain.LocationTypePallet, &domain.Capacity{MaxWeight: 1000, MaxVolume: 2.0, MaxQty: 72}},
		{"B-02-01-01", "Z-STO-B", domain.LocationTypeShelf, &domain.Capacity{MaxWeight: 500, MaxVolume: 1.0, MaxQty: 200}},
		{"B-02-01-02", "Z-STO-B", domain.LocationTypeShelf, &domain.Capacity{MaxWeight: 500, MaxVolume: 1.0, MaxQty: 200}},

		// Picking zone — flow racks
		{"PICK-A01", "Z-PICK", domain.LocationTypeShelf, &domain.Capacity{MaxWeight: 300, MaxVolume: 0.5, MaxQty: 100}},
		{"PICK-A02", "Z-PICK", domain.LocationTypeShelf, &domain.Capacity{MaxWeight: 300, MaxVolume: 0.5, MaxQty: 100}},
		{"PICK-A03", "Z-PICK", domain.LocationTypeShelf, &domain.Capacity{MaxWeight: 300, MaxVolume: 0.5, MaxQty: 100}},
		{"PICK-B01", "Z-PICK", domain.LocationTypeShelf, &domain.Capacity{MaxWeight: 300, MaxVolume: 0.5, MaxQty: 100}},
		{"PICK-B02", "Z-PICK", domain.LocationTypeShelf, &domain.Capacity{MaxWeight: 300, MaxVolume: 0.5, MaxQty: 100}},
		{"PICK-B03", "Z-PICK", domain.LocationTypeShelf, &domain.Capacity{MaxWeight: 300, MaxVolume: 0.5, MaxQty: 100}},

		// Shipping docks
		{"SHIP-DOCK-01", "Z-SHIP", domain.LocationTypeFloor, nil},
		{"SHIP-DOCK-02", "Z-SHIP", domain.LocationTypeFloor, nil},
		{"SHIP-DOCK-03", "Z-SHIP", domain.LocationTypeFloor, nil},

		// Returns area
		{"RET-01", "Z-RET", domain.LocationTypeFloor, nil},
		{"RET-02", "Z-RET", domain.LocationTypeFloor, nil},
	}

	// Build a zone lookup by code.
	zoneByCode := make(map[string]*domain.Zone)
	for _, z := range zones {
		zoneByCode[z.Code] = z
	}

	locations := make(map[string]*domain.Location)
	for _, ld := range defs {
		z, ok := zoneByCode[ld.ZoneCode]
		if !ok {
			fmt.Fprintf(os.Stderr, "Zone %s not found for location %s\n", ld.ZoneCode, ld.Code)
			os.Exit(1)
		}
		loc := &domain.Location{
			ZoneID:       z.ID,
			WarehouseID:  warehouseID,
			Code:         ld.Code,
			Barcode:      fmt.Sprintf("LOC-%s", ld.Code),
			LocationType: ld.LocationType,
			Capacity:     ld.Capacity,
			Status:       domain.LocationStatusEmpty,
		}
		if err := repo.CreateLocation(ctx, loc); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create location %s: %v\n", ld.Code, err)
			os.Exit(1)
		}
		locations[ld.Code] = loc
	}
	log.Info(fmt.Sprintf("  ✓ Created %d locations across %d zones", len(defs), len(zones)))
	return locations
}

// skuDef defines a SKU to seed.
type skuDef struct {
	Code        string
	Name        string
	Description string
	Barcode     string
	Category    string
	UOM         domain.UOM
	Attributes  domain.Attributes
}

func seedSKUs(ctx context.Context, repo *postgres.InventoryRepo, log *logger.Logger) []*domain.SKU {
	defs := []skuDef{
		{
			Code: "SKU-10001", Name: "USB-C Cable 1m", Description: "USB-C to USB-C cable, 1 meter, braided",
			Barcode: "BAR-10001", Category: "Electronics",
			UOM: domain.UOM{BaseUnit: "EA", Weight: 0.05, Length: 1.0, Width: 0.01, Height: 0.01},
			Attributes: domain.Attributes{"color": "black", "connector": "USB-C", "length_cm": "100"},
		},
		{
			Code: "SKU-10002", Name: "USB-C Cable 2m", Description: "USB-C to USB-C cable, 2 meters, braided",
			Barcode: "BAR-10002", Category: "Electronics",
			UOM: domain.UOM{BaseUnit: "EA", Weight: 0.09, Length: 2.0, Width: 0.01, Height: 0.01},
			Attributes: domain.Attributes{"color": "black", "connector": "USB-C", "length_cm": "200"},
		},
		{
			Code: "SKU-10003", Name: "Wireless Mouse", Description: "2.4GHz wireless optical mouse",
			Barcode: "BAR-10003", Category: "Electronics",
			UOM: domain.UOM{BaseUnit: "EA", Weight: 0.12, Length: 0.11, Width: 0.06, Height: 0.04},
			Attributes: domain.Attributes{"color": "dark gray", "dpi": "1600", "battery": "AA"},
		},
		{
			Code: "SKU-10004", Name: "Mechanical Keyboard", Description: "Full-size mechanical keyboard, blue switches",
			Barcode: "BAR-10004", Category: "Electronics",
			UOM: domain.UOM{BaseUnit: "EA", Weight: 0.95, Length: 0.44, Width: 0.14, Height: 0.04},
			Attributes: domain.Attributes{"color": "black", "switch": "blue", "layout": "US-104"},
		},
		{
			Code: "SKU-10005", Name: "27-inch Monitor", Description: "27-inch 4K IPS monitor, HDMI+DP",
			Barcode: "BAR-10005", Category: "Electronics",
			UOM: domain.UOM{BaseUnit: "EA", Weight: 5.6, Length: 0.62, Width: 0.20, Height: 0.42},
			Attributes: domain.Attributes{"size": "27", "resolution": "3840x2160", "panel": "IPS"},
		},
		{
			Code: "SKU-20001", Name: "Cotton T-Shirt White M", Description: "Men's plain white cotton t-shirt, size M",
			Barcode: "BAR-20001", Category: "Apparel",
			UOM: domain.UOM{BaseUnit: "EA", Weight: 0.20, Length: 0.30, Width: 0.25, Height: 0.02},
			Attributes: domain.Attributes{"color": "white", "size": "M", "material": "100% cotton"},
		},
		{
			Code: "SKU-20002", Name: "Cotton T-Shirt White L", Description: "Men's plain white cotton t-shirt, size L",
			Barcode: "BAR-20002", Category: "Apparel",
			UOM: domain.UOM{BaseUnit: "EA", Weight: 0.22, Length: 0.32, Width: 0.27, Height: 0.02},
			Attributes: domain.Attributes{"color": "white", "size": "L", "material": "100% cotton"},
		},
		{
			Code: "SKU-20003", Name: "Denim Jeans Blue 32", Description: "Slim-fit denim jeans, blue, waist 32",
			Barcode: "BAR-20003", Category: "Apparel",
			UOM: domain.UOM{BaseUnit: "EA", Weight: 0.65, Length: 0.40, Width: 0.30, Height: 0.04},
			Attributes: domain.Attributes{"color": "blue", "size": "32", "fit": "slim"},
		},
		{
			Code: "SKU-30001", Name: "Bottled Water 500ml x24", Description: "Mineral water, 500ml bottles, pack of 24",
			Barcode: "BAR-30001", Category: "Beverages",
			UOM: domain.UOM{BaseUnit: "CS", PackUnit: "BTL", PackQty: 24, Weight: 12.5, Length: 0.40, Width: 0.27, Height: 0.22},
			Attributes: domain.Attributes{"volume_ml": "500", "pack_size": "24", "shelf_life_days": "365"},
		},
		{
			Code: "SKU-30002", Name: "Green Tea Box 20pk", Description: "Premium green tea bags, box of 20",
			Barcode: "BAR-30002", Category: "Beverages",
			UOM: domain.UOM{BaseUnit: "EA", Weight: 0.10, Length: 0.14, Width: 0.07, Height: 0.07},
			Attributes: domain.Attributes{"type": "green tea", "count": "20", "origin": "Hangzhou"},
		},
		{
			Code: "SKU-40001", Name: "A4 Printer Paper", Description: "A4 80gsm copy paper, 500 sheets per ream",
			Barcode: "BAR-40001", Category: "Office Supplies",
			UOM: domain.UOM{BaseUnit: "REAM", Weight: 2.5, Length: 0.30, Width: 0.21, Height: 0.05},
			Attributes: domain.Attributes{"size": "A4", "gsm": "80", "sheets": "500", "brightness": "92"},
		},
		{
			Code: "SKU-40002", Name: "Ballpoint Pen Blue x12", Description: "Medium point blue ballpoint pens, box of 12",
			Barcode: "BAR-40002", Category: "Office Supplies",
			UOM: domain.UOM{BaseUnit: "BOX", PackUnit: "EA", PackQty: 12, Weight: 0.12, Length: 0.15, Width: 0.05, Height: 0.02},
			Attributes: domain.Attributes{"color": "blue", "tip": "medium", "count": "12"},
		},
		{
			Code: "SKU-50001", Name: "All-Purpose Cleaner 5L", Description: "Multi-surface cleaning solution, 5 liter jug",
			Barcode: "BAR-50001", Category: "Cleaning",
			UOM: domain.UOM{BaseUnit: "EA", Weight: 5.2, Length: 0.18, Width: 0.12, Height: 0.30},
			Attributes: domain.Attributes{"volume_l": "5", "scent": "lemon", "type": "all-purpose"},
		},
		{
			Code: "SKU-50002", Name: "Disposable Gloves L x100", Description: "Nitrile disposable gloves, size L, box of 100",
			Barcode: "BAR-50002", Category: "Cleaning",
			UOM: domain.UOM{BaseUnit: "BOX", PackUnit: "EA", PackQty: 100, Weight: 0.50, Length: 0.22, Width: 0.12, Height: 0.06},
			Attributes: domain.Attributes{"size": "L", "material": "nitrile", "count": "100", "color": "blue"},
		},
		{
			Code: "SKU-60001", Name: "Smartphone Case iPhone 16", Description: "Shockproof TPU case for iPhone 16",
			Barcode: "BAR-60001", Category: "Accessories",
			UOM: domain.UOM{BaseUnit: "EA", Weight: 0.03, Length: 0.16, Width: 0.08, Height: 0.01},
			Attributes: domain.Attributes{"compatible": "iPhone 16", "material": "TPU", "color": "clear"},
		},
		{
			Code: "SKU-60002", Name: "Screen Protector Tempered Glass", Description: "9H tempered glass screen protector, universal fit",
			Barcode: "BAR-60002", Category: "Accessories",
			UOM: domain.UOM{BaseUnit: "EA", Weight: 0.02, Length: 0.18, Width: 0.10, Height: 0.001},
			Attributes: domain.Attributes{"hardness": "9H", "thickness_mm": "0.33", "type": "tempered glass"},
		},
	}

	var skus []*domain.SKU
	for _, sd := range defs {
		s := &domain.SKU{
			Code:        sd.Code,
			Name:        sd.Name,
			Description: sd.Description,
			Barcode:     sd.Barcode,
			Category:    sd.Category,
			UOM:         sd.UOM,
			Attributes:  sd.Attributes,
			Status:      domain.SKUStatusActive,
		}
		if err := repo.CreateSKU(ctx, s); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create SKU %s: %v\n", sd.Code, err)
			os.Exit(1)
		}
		skus = append(skus, s)
	}
	log.Info(fmt.Sprintf("  ✓ Created %d SKUs across %d categories", len(skus), uniqueCategories(defs)))
	return skus
}

func seedInventory(ctx context.Context, repo *postgres.InventoryRepo, warehouseID uuid.UUID, skus []*domain.SKU, locations map[string]*domain.Location, log *logger.Logger) {
	// Map SKUs to storage/pick locations with quantities.
	type invEntry struct {
		SKUIdx    int
		LocCode   string
		Qty       float64
		BatchNo   string
		Status    domain.InventoryStatus
	}

	entries := []invEntry{
		// USB-C cables in Storage A
		{0, "A-01-01-01", 500, "", domain.InventoryStatusAvailable},
		{1, "A-01-01-02", 300, "", domain.InventoryStatusAvailable},
		// Wireless mouse
		{2, "A-01-02-01", 150, "", domain.InventoryStatusAvailable},
		// Mechanical keyboard
		{3, "A-01-02-02", 80, "", domain.InventoryStatusAvailable},
		// Monitor (big items, low qty)
		{4, "A-01-03-01", 25, "", domain.InventoryStatusAvailable},
		{4, "A-01-03-02", 10, "", domain.InventoryStatusAvailable},
		// T-shirts in shelves
		{5, "A-02-01-01", 200, "LOT-TS-001", domain.InventoryStatusAvailable},
		{6, "A-02-01-02", 180, "LOT-TS-001", domain.InventoryStatusAvailable},
		// Jeans
		{7, "A-02-01-03", 100, "", domain.InventoryStatusAvailable},
		// Beverages in Storage B
		{8, "B-01-01-01", 50, "", domain.InventoryStatusAvailable},
		{9, "B-01-01-02", 120, "", domain.InventoryStatusAvailable},
		// Office supplies in Storage B
		{10, "B-01-02-01", 80, "", domain.InventoryStatusAvailable},
		{11, "B-01-02-02", 200, "", domain.InventoryStatusAvailable},
		// Cleaning products in shelves
		{12, "B-02-01-01", 30, "", domain.InventoryStatusAvailable},
		{13, "B-02-01-02", 60, "", domain.InventoryStatusAvailable},
		// Accessories in shelves
		{14, "A-02-01-04", 250, "", domain.InventoryStatusAvailable},
		{15, "B-02-01-01", 300, "", domain.InventoryStatusAvailable},
		// Pick face inventory (small quantities for picking)
		{0, "PICK-A01", 20, "", domain.InventoryStatusAvailable},
		{1, "PICK-A02", 15, "", domain.InventoryStatusAvailable},
		{2, "PICK-A03", 10, "", domain.InventoryStatusAvailable},
		{5, "PICK-B01", 30, "LOT-TS-001", domain.InventoryStatusAvailable},
		{6, "PICK-B02", 25, "LOT-TS-001", domain.InventoryStatusAvailable},
		{10, "PICK-B03", 15, "", domain.InventoryStatusAvailable},
	}

	created := 0
	for _, e := range entries {
		if e.SKUIdx >= len(skus) {
			continue
		}
		loc, ok := locations[e.LocCode]
		if !ok {
			fmt.Fprintf(os.Stderr, "Location %s not found\n", e.LocCode)
			os.Exit(1)
		}
		sku := skus[e.SKUIdx]
		inv := &domain.Inventory{
			SKUID:       sku.ID,
			LocationID:  loc.ID,
			WarehouseID: warehouseID,
			BatchNo:     e.BatchNo,
			Qty:         e.Qty,
			ReservedQty: 0,
			Status:      e.Status,
		}
		if err := repo.CreateInventory(ctx, inv); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create inventory for SKU %s at %s: %v\n", sku.Code, e.LocCode, err)
			continue
		}
		created++
	}
	log.Info(fmt.Sprintf("  ✓ Created %d inventory records", created))
}

func seedAdminUser(ctx context.Context, db *postgres.DB, log *logger.Logger) {
	userRepo := postgres.NewUserRepo(db)

	// ── 1. Default Roles (idempotent) ──────────────────────────
	roleCount, err := userRepo.CountRoles(ctx)
	if err != nil {
		log.Info(fmt.Sprintf("  ⚠ Could not count roles: %v", err))
		return
	}

	if roleCount == 0 {
		defaultRoles := []struct {
			name        string
			description string
			permissions []domain.Permission
		}{
			{"admin", "System Administrator", []domain.Permission{{Resource: "*", Actions: []string{"*"}}}},
			{"operator", "Warehouse Operator", []domain.Permission{
				{Resource: "warehouse", Actions: []string{"read"}},
				{Resource: "inventory", Actions: []string{"read", "update"}},
				{Resource: "order", Actions: []string{"read", "create", "update"}},
				{Resource: "task", Actions: []string{"read", "update"}},
			}},
			{"picker", "Picker (PDA User)", []domain.Permission{
				{Resource: "task", Actions: []string{"read", "update"}},
				{Resource: "inventory", Actions: []string{"read"}},
			}},
		}
		for _, r := range defaultRoles {
			role := &domain.Role{
				Name:        r.name,
				Description: r.description,
				Permissions: r.permissions,
			}
			if err := userRepo.CreateRole(ctx, role); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create role %s: %v\n", r.name, err)
				os.Exit(1)
			}
		}
		log.Info(fmt.Sprintf("  ✓ Created %d default roles", len(defaultRoles)))
	}

	// ── 2. Admin User (idempotent) ─────────────────────────────
	adminUser, err := userRepo.GetUserByUsername(ctx, "admin")
	if err != nil {
		// Admin user does not exist — create one.
		hash, hashErr := service.HashPassword("admin123")
		if hashErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to hash password: %v\n", hashErr)
			os.Exit(1)
		}

		// Fetch the admin role ID.
		roles, roleErr := userRepo.ListRoles(ctx)
		if roleErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to list roles: %v\n", roleErr)
			os.Exit(1)
		}

		var adminRoleID uuid.UUID
		for _, r := range roles {
			if r.Name == "admin" {
				adminRoleID = r.ID
				break
			}
		}
		if adminRoleID == uuid.Nil {
			fmt.Fprintf(os.Stderr, "Admin role not found after seeding\n")
			os.Exit(1)
		}

		admin := &domain.User{
			Username:     "admin",
			Email:        "admin@wms.local",
			PasswordHash: hash,
			DisplayName:  "System Admin",
			RoleIDs:      []uuid.UUID{adminRoleID},
			Status:       domain.UserStatusActive,
		}
		if createErr := userRepo.CreateUser(ctx, admin); createErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to create admin user: %v\n", createErr)
			os.Exit(1)
		}
		log.Info("  ✓ Admin user created with hashed password (admin123)")
		return
	}

	// Admin user exists — update password hash if it looks like a placeholder.
	if len(adminUser.PasswordHash) < 20 || adminUser.PasswordHash[:4] != "$2a$" {
		hash, hashErr := service.HashPassword("admin123")
		if hashErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to hash password: %v\n", hashErr)
			os.Exit(1)
		}
		if _, execErr := db.Pool.Exec(ctx, `UPDATE users SET password_hash=$1 WHERE username='admin'`, hash); execErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to update admin password: %v\n", execErr)
			os.Exit(1)
		}
		log.Info("  ✓ Admin user password updated to bcrypt hash")
	} else {
		log.Info("  ✓ Admin user already exists with valid password hash")
	}
}

// ── Helpers ────────────────────────────────────────────────────────────────────

func uniqueCategories(defs []skuDef) int {
	seen := make(map[string]bool)
	for _, d := range defs {
		seen[d.Category] = true
	}
	return len(seen)
}

func totalLocations() int {
	// We hard-code the location count since zones don't track their locations inline.
	return 35
}
