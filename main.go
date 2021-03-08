package main

import (
	"database/sql"
	"fmt"
	echoSwagger "github.com/swaggo/echo-swagger"
	"math/big"
	"net"
	"strconv"

	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoLog "github.com/labstack/gommon/log"
	"github.com/neko-neko/echo-logrus/v2/log"
	_ "github.com/ryuhon/ad-server/docs"
	"github.com/sirupsen/logrus"
)

func initDB() (db *sqlx.DB) {

	dbDriver := "mysql"
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	db, err := sqlx.Open(dbDriver, dbUser+":"+dbPass+"@tcp("+dbHost+":"+dbPort+")/"+dbName)

	if err != nil {
		panic(err)
	}

	if db == nil {
		panic(err)
	}

	return db
}


// AD 광고
type Ad struct {
	Aid        				int    `json:"aid" db:"aid"`
	Mid						int		`json:"mid" db:"mid"`
	Title        			string    `json:"title" db:"title"`
	BannerSize       		int 	`json:"bannerSize" db:"banner_size"`
	BannerUrl    			string `json:"bannerUrl" db:"banner_url"`
	ImpressionTrackingUrl  	string `json:"impressionTrackingUrl" db:"impression_tracking_url"`
	RedirectUrl    			string `json:"redirectUrl" db:"redirect_url"`

}

type UserLog struct {
	RegDate    				string `json:"regDate" db:"reg_date"`
	Aid        				int    `json:"aid" db:"aid"`
	Url						int		`json:"url" db:"url"`
	Ip4        				string    `json:"ip4" db:"ip4"`

}


type Error struct {
	Message string `string:"message"`
}

const EmptyAd = "광고가 없습니다."

// adGet 광고 조회
// @Summary 광고 조회
// @Description 하나의 광고를 가져온다
// @Tags 광고
// @Accept json
// @Produce json
// @Success 200 {object} Ad
// @Router /api/ad [get]
func adGet(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {

		result, err := dbGetAd(db )

		ip := c.RealIP()
		dbSaveRequest(db,result.Aid,"",ip)

		if err == nil {
			return c.JSON(http.StatusOK, result)
		} else {
			errResult := Error{}
			if err == sql.ErrNoRows {
				errResult.Message = EmptyAd
				return c.JSON(http.StatusNotFound, errResult)
			} else {
				errResult.Message = err.Error()
				return c.JSON(http.StatusBadGateway, errResult)
			}
		}
	}
}


//
// @Summary  로깅
// @Description  로깅
// @Tags 로깅
// @Param action path string true "request , click , impression "
// @Param aid path int true "Aid"
// @Accept json
// @Produce json
// @Success 200 {object} Ad
// @Router /api/log/{action}/{aid} [get]
func loggingAd(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		aid, _ := strconv.Atoi(c.Param("aid"))
		action  :=  c.Param("action")
		ip := c.RealIP()
		var err error
		if  action == "click"  {
			err = dbSaveClick(db,aid,"",ip )
		} else if action == "request" {
			err = dbSaveRequest(db,aid,"",ip )
		} else if action == "impression" {
			err = dbSaveImpression(db,aid,"",ip )
		}


		if err == nil {
 			return c.JSON(http.StatusOK,"")
		} else {

			return c.JSON(http.StatusBadRequest, "")
		}
	}
}

func dbGetAd(db *sqlx.DB) (Ad, error) {
	query := `SELECT 
				aid,
				mid,
				title,
				banner_size,
				banner_url,
				impression_tracking_url,
				redirect_url
			FROM 
			ad 
			WHERE platform_type = 31 
			ORDER BY RAND() LIMIT 1 `
	result := Ad{}
 	err := db.Get(&result,query)
	fmt.Printf("%#v\n", result)

	return result, err
}
func Inet_Aton(ip net.IP) int64 {
	ipv4Int := big.NewInt(0)
	ipv4Int.SetBytes(ip.To4())
	return ipv4Int.Int64()
}
func dbSaveClick(db *sqlx.DB, aid int , url string, ip string  ) (  error) {
	query := `INSERT INTO click (
				aid,
				url,
				ip4 
        	) VALUES( 
				?,
				?, 
				?
        )`

	_, err := db.Exec(query,
		aid,
		url ,
		ip )
	fmt.Printf("%#v\n", err)
 	return   err
}
func dbSaveRequest(db *sqlx.DB, aid int , url string, ip string  ) (  error) {
	query := `INSERT INTO request (
				aid,
				url,
				ip4 
        	) VALUES( 
				?,
				?, 
				?
        )`

	_, err := db.Exec(query,
		aid,
		url ,
		ip )
	fmt.Printf("%#v\n", err)
	return   err
}
func dbSaveImpression(db *sqlx.DB, aid int , url string, ip string  ) (  error) {
	query := `INSERT INTO impression (
				aid,
				url,
				ip4 
        	) VALUES( 
				?,
				?, 
				?
        )`

	_, err := db.Exec(query,
		aid,
		url ,
		ip )
	fmt.Printf("%#v\n", err)
	return   err
}


// @title Ad Server
// @description 광고 딜리버리 서법
// @version 1.0
// @host localhost:80
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	e := echo.New()

	db := initDB()
	//migrage(db)

	log.Logger().SetOutput(os.Stdout)
	log.Logger().SetLevel(echoLog.INFO)
	log.Logger().SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})
	e.Logger = log.Logger()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	//e.Logger.SetLevel(99)
	// CORS
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		//AllowHeaders:     []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
		AllowCredentials: true,
	}))

	// Router

	e.GET("/api/ad", adGet(db))
	e.GET("/api/log/:action/:aid", loggingAd(db))
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.File("/test", "public/index.html")

	e.Logger.Fatal(e.Start(":80"))
}
