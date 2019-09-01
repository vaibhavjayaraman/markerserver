package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/worldhistorymap/backend/pkg/shared"
	"go.uber.org/zap"
)

type LatLonReq struct {
	Lat          float64 `json:"lat"`
	Lon          float64 `json:"lon"`
	Year         int     `json:"year"`
	FileReqLimit int     `json:"file_request_size"`
}

type Server struct {
	mux    *http.ServeMux
	db     *sql.DB
	query  string
	logger *zap.Logger
	config *shared.Config
}

func NewServer(config *shared.Config, logger *zap.Logger) (*Server, error) {
	query := "SELECT url, info, title, source, lat, lon FROM markers WHERE beg_year <= $1 AND end_year >= $1" +
		"ORDER BY geom <-> ST_SetSRID(ST_MakePoint($2, $3), 4326) LIMIT $4;"
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	return &Server{
		mux:   mux,
		db:    db,
		query: query,
	}, nil
}

func (s *Server) Run() error {
	s.mux.HandleFunc("/articles", s.articles)
	return http.ListenAndServe(":8000", s.mux)
}

func (s *Server) articles(w http.ResponseWriter, r *http.Request) {
	var llrq LatLonReq
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	err = json.Unmarshal(body, &llrq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	queryStmt, err := s.db.Prepare(s.query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	markers, err := getArticleMarkers(queryStmt, &llrq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	markerJson, err := json.Marshal(markers)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(markerJson)
	w.WriteHeader(http.StatusOK)
}

func getArticleMarkers(query *sql.Stmt, llrq *LatLonReq) ([]shared.Marker, error) {
	var markers []shared.Marker
	rows, err := query.Query(llrq.Year, llrq.Lon, llrq.Lat, llrq.FileReqLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var marker shared.Marker
		if err = rows.Scan(&marker); err != nil {
			return nil, err
		}
		markers = append(markers, marker)
	}
	return markers, nil
}
