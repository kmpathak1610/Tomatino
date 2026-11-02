package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://foodapp:foodapp@localhost:5432/fooddelivery?sslmode=disable"
	}
	conn, _ := pgx.Connect(context.Background(), dbURL)
	defer conn.Close(context.Background())

	// city
	cityID := uuid.New().String()
	conn.Exec(context.Background(), "INSERT INTO cities(id,name,state) VALUES($1,'Bengaluru','Karnataka') ON CONFLICT DO NOTHING", cityID)
	// Actually ensure one city
	var existingCity string
	conn.QueryRow(context.Background(), "SELECT id FROM cities LIMIT 1").Scan(&existingCity)
	if existingCity == "" {
		existingCity = cityID
		conn.Exec(context.Background(), "INSERT INTO cities(id,name,state) VALUES($1,'Bengaluru','Karnataka')", existingCity)
	}

	// user owner (staff login: owner@spice.com / password123)
	ownerID := uuid.New().String()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	conn.Exec(context.Background(), `INSERT INTO users(id,phone,email,password_hash,status)
		VALUES($1,'+919000000000','owner@spice.com',$2,'ACTIVE')
		ON CONFLICT (email) DO UPDATE SET password_hash=EXCLUDED.password_hash`, ownerID, string(hash))
	conn.Exec(context.Background(), `UPDATE users SET email='owner@spice.com', password_hash=$1, status='ACTIVE'
		WHERE phone='+919000000000'`, string(hash))
	conn.QueryRow(context.Background(), "SELECT id FROM users WHERE email='owner@spice.com'").Scan(&ownerID)
	conn.Exec(context.Background(), "INSERT INTO user_roles(user_id,role) VALUES($1,'RESTAURANT_OWNER') ON CONFLICT DO NOTHING", ownerID)

	// admin account: admin@foodexpress.com / password123
	adminID := uuid.New().String()
	conn.Exec(context.Background(), `INSERT INTO users(id,phone,email,password_hash,status)
		VALUES($1,'+919000000001','admin@foodexpress.com',$2,'ACTIVE')
		ON CONFLICT (email) DO UPDATE SET password_hash=EXCLUDED.password_hash`, adminID, string(hash))
	conn.QueryRow(context.Background(), "SELECT id FROM users WHERE email='admin@foodexpress.com'").Scan(&adminID)
	conn.Exec(context.Background(), "INSERT INTO user_roles(user_id,role) VALUES($1,'SUPER_ADMIN') ON CONFLICT DO NOTHING", adminID)

	// restaurant
	restID := uuid.New().String()
	conn.Exec(context.Background(), `INSERT INTO restaurants(id,owner_user_id,name,slug,description,cuisine_types,status,avg_rating,total_ratings)
		VALUES($1,$2,'Spice Garden','spice-garden','Authentic Indian cuisine',ARRAY['Indian','North Indian'],'APPROVED',4.5,120) ON CONFLICT DO NOTHING`,
		restID, ownerID)
	conn.Exec(context.Background(), "UPDATE restaurants SET owner_user_id=$1 WHERE slug='spice-garden'", ownerID)
	conn.QueryRow(context.Background(), "SELECT id FROM restaurants WHERE slug='spice-garden'").Scan(&restID)

	// branch
	branchID := uuid.New().String()
	conn.Exec(context.Background(), `INSERT INTO restaurant_branches(id,restaurant_id,name,address_line1,city_id,location,is_open,min_order_amount,delivery_radius_km)
		VALUES($1,$2,'Main Branch','MG Road', (SELECT id FROM cities LIMIT 1), ST_SetSRID(ST_MakePoint(77.5946,12.9716),4326)::geography, TRUE, 100, 8) ON CONFLICT DO NOTHING`,
		branchID, restID)

	// menu categories
	catID := uuid.New().String()
	conn.Exec(context.Background(), "INSERT INTO menu_categories(id,restaurant_id,name,sort_order) VALUES($1,$2,'Main Course',1) ON CONFLICT DO NOTHING", catID, restID)

	// menu items
	items := []struct{ name string; price float64; veg bool }{
		{"Butter Chicken", 280, false},
		{"Paneer Tikka", 220, true},
		{"Dal Makhani", 180, true},
		{"Biryani", 320, false},
	}
	for _, it := range items {
		id := uuid.New().String()
		conn.Exec(context.Background(), `INSERT INTO menu_items(id,restaurant_id,category_id,name,base_price,is_veg,is_available,is_bestseller,prep_time_min)
			VALUES($1,$2,$3,$4,$5,$6,TRUE,FALSE,20) ON CONFLICT DO NOTHING`, id, restID, catID, it.name, it.price, it.veg)
	}

	// coupon
	conn.Exec(context.Background(), `INSERT INTO coupons(code,description,discount_type,discount_value,min_order_amount,max_discount,usage_limit,per_user_limit,valid_from,valid_until,is_active)
		VALUES('WELCOME50','50% off up to 100','PERCENTAGE',50,200,100,1000,1,now(),now()+interval '365 days',TRUE) ON CONFLICT(code) DO NOTHING`)

	// delivery partner: rider@foodexpress.com / password123
	riderID := uuid.New().String()
	conn.Exec(context.Background(), `INSERT INTO users(id,phone,email,password_hash,status)
		VALUES($1,'+919000000002','rider@foodexpress.com',$2,'ACTIVE')
		ON CONFLICT (email) DO UPDATE SET password_hash=EXCLUDED.password_hash`, riderID, string(hash))
	conn.QueryRow(context.Background(), "SELECT id FROM users WHERE email='rider@foodexpress.com'").Scan(&riderID)
	conn.Exec(context.Background(), "INSERT INTO user_roles(user_id,role) VALUES($1,'DELIVERY_PARTNER') ON CONFLICT DO NOTHING", riderID)
	conn.Exec(context.Background(), `INSERT INTO delivery_partners(id,user_id,vehicle_type,status)
		VALUES($1,$2,'BIKE','APPROVED') ON CONFLICT (user_id) DO NOTHING`, uuid.New().String(), riderID)

	fmt.Println("seeded")
}