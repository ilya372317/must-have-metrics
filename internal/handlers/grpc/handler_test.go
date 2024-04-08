package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"sort"
	"sync"
	"testing"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/service"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	pb "github.com/ilya372317/must-have-metrics/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

var (
	lis    *bufconn.Listener
	strg   *storage.InMemoryStorage
	client pb.MetricsServiceClient
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	clientConn, err := grpc.DialContext(ctx,
		"bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	client = pb.NewMetricsServiceClient(clientConn)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = clientConn.Close()
		if err != nil {
			log.Printf("failed close grpc client conn: %v", err)
		}
	}()
	cnfg := config.ServerConfig{
		StoreInterval: 1,
	}
	strg = storage.NewInMemoryStorage()
	serv := service.NewMetricsService(strg, &cnfg)
	s := grpc.NewServer()
	pb.RegisterMetricsServiceServer(s, New(serv, &cnfg))
	lis = bufconn.Listen(1024 * 1024)

	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.Serve(lis); err != nil {
			log.Printf("failed run grpc server: %v\n", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		s.GracefulStop()
	}()

	m.Run()
	cancel()
	wg.Wait()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	conn, err := lis.Dial()
	if err != nil {
		return nil, fmt.Errorf("failed make test connection for grpc: %w", err)
	}

	return conn, nil
}

func TestServer_Index(t *testing.T) {
	tests := []struct {
		name string
		data []entity.Alert
		want []*pb.Metrics
	}{
		{
			name: "success case with empty storage",
			data: nil,
			want: nil,
		},
		{
			name: "filled storage case",
			data: []entity.Alert{
				{
					IntValue: intPointer(1),
					Type:     "counter",
					Name:     "alert1",
				},
				{
					FloatValue: floatPointer(1.1),
					Type:       "gauge",
					Name:       "alert2",
				},
			},
			want: []*pb.Metrics{
				{
					Delta: intPointer(1),
					Id:    "alert1",
					Type:  "counter",
				},
				{
					Value: floatPointer(1.1),
					Id:    "alert2",
					Type:  "gauge",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			for _, item := range tt.data {
				err := strg.Save(ctx, item.Name, item)
				require.NoError(t, err)
			}
			resp, err := client.Index(ctx, &pb.IndexMetricsRequest{})
			require.NoError(t, err)
			got := resp.Metrics
			sort.SliceStable(tt.want, func(i, j int) bool {
				return tt.want[i].Id < tt.want[j].Id
			})
			sort.SliceStable(got, func(i, j int) bool {
				return got[i].Id < got[j].Id
			})
			for _, w := range tt.want {
				g := got[0]
				got = got[1:]
				assert.Equal(t, w.Id, g.Id)
				assert.Equal(t, w.Type, g.Type)
				if w.Delta != nil {
					assert.Equal(t, *w.Delta, *g.Delta)
				}
				if w.Value != nil {
					assert.Equal(t, *w.Value, *g.Value)
				}
			}
			strg.Reset()
		})
	}
}

func TestServer_BulkUpdate(t *testing.T) {
	tests := []struct {
		name        string
		requestData []*pb.Metrics
		want        []*pb.Metrics
		errCode     codes.Code
		wantErr     bool
	}{
		{
			name: "success filled case",
			requestData: []*pb.Metrics{
				{
					Value: floatPointer(1.1),
					Id:    "alert1",
					Type:  "gauge",
				},
				{
					Delta: intPointer(1),
					Id:    "alert2",
					Type:  "counter",
				},
			},
			wantErr: false,
			want: []*pb.Metrics{
				{
					Value: floatPointer(1.1),
					Id:    "alert1",
					Type:  "gauge",
				},
				{
					Delta: intPointer(1),
					Id:    "alert2",
					Type:  "counter",
				},
			},
		},
		{
			name:        "success empty case",
			requestData: []*pb.Metrics{},
			wantErr:     false,
			want:        nil,
		},
		{
			name: "invalid type case",
			requestData: []*pb.Metrics{
				{
					Value: floatPointer(1.2),
					Id:    "alert1",
					Type:  "invalid-type",
				},
			},
			wantErr: true,
			want:    nil,
			errCode: codes.InvalidArgument,
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.BulkUpdate(ctx, &pb.BulkUpdateMetricsRequest{Metrics: tt.requestData})
			if tt.wantErr {
				e, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tt.errCode, e.Code())
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			assert.EqualValues(t, tt.want, resp.Metrics)
			strg.Reset()
		})
	}
}

func TestServer_Show(t *testing.T) {
	type argument struct {
		ID   string
		Type string
	}
	type want struct {
		ID    string
		Type  string
		Delta *int64
		Value *float64
	}
	tests := []struct {
		data    []entity.Alert
		name    string
		arg     argument
		want    want
		errCode codes.Code
		wantErr bool
	}{
		{
			name: "invalid type case",
			arg: argument{
				ID:   "alert1",
				Type: "invalid-type",
			},
			errCode: codes.InvalidArgument,
			wantErr: true,
		},
		{
			name: "not found case with empty storage",
			arg: argument{
				ID:   "alert1",
				Type: "gauge",
			},
			errCode: codes.NotFound,
			wantErr: true,
		},
		{
			data: []entity.Alert{
				{
					IntValue: intPointer(1),
					Type:     "counter",
					Name:     "alert2",
				},
				{
					FloatValue: floatPointer(1.1),
					Type:       "gauge",
					Name:       "alert3",
				},
			},
			name: "not found case with filled storage",
			arg: argument{
				ID:   "alert1",
				Type: "gauge",
			},
			errCode: codes.NotFound,
			wantErr: true,
		},
		{
			data: []entity.Alert{
				{
					IntValue: intPointer(1),
					Type:     "counter",
					Name:     "alert2",
				},
				{
					FloatValue: floatPointer(1.1),
					Type:       "gauge",
					Name:       "alert3",
				},
			},
			name: "success case with counter metric",
			arg: argument{
				ID:   "alert2",
				Type: "counter",
			},
			want: want{
				ID:    "alert2",
				Type:  "counter",
				Delta: intPointer(1),
			},
			wantErr: false,
		},
		{
			data: []entity.Alert{
				{
					IntValue: intPointer(1),
					Type:     "counter",
					Name:     "alert2",
				},
				{
					FloatValue: floatPointer(1.1),
					Type:       "gauge",
					Name:       "alert3",
				},
			},
			name: "success case with gauge metric",
			arg: argument{
				ID:   "alert3",
				Type: "gauge",
			},
			want: want{
				ID:    "alert3",
				Type:  "gauge",
				Value: floatPointer(1.1),
			},
			wantErr: false,
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, d := range tt.data {
				err := strg.Save(context.Background(), d.Name, d)
				require.NoError(t, err)
			}

			in := &pb.ShowMetricsRequest{
				Id:   tt.arg.ID,
				Type: tt.arg.Type,
			}
			resp, err := client.Show(ctx, in)
			if tt.wantErr {
				require.Error(t, err)
				e, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, e.Code())
				return
			} else {
				require.NoError(t, err)
			}

			got := resp.Metrics
			assert.Equal(t, tt.want.ID, got.Id)
			assert.Equal(t, tt.want.Type, got.Type)

			if tt.want.Value == nil {
				assert.Nil(t, got.Value)
			} else {
				if got.Value == nil {
					t.Errorf("value of field `Value` expected to be not nil.")
					return
				}
				assert.Equal(t, *tt.want.Value, *got.Value)
			}

			if tt.want.Delta == nil {
				assert.Nil(t, got.Delta)
			} else {
				if got.Delta == nil {
					t.Errorf("value of field `Delta` expected to be not nil")
					return
				}
				assert.Equal(t, *tt.want.Delta, *got.Delta)
			}
			strg.Reset()
		})
	}
}

func TestServer_Update(t *testing.T) {
	type want struct {
		ID    string
		Type  string
		Delta *int64
		Value *float64
	}
	type argument struct {
		ID    string
		Type  string
		Delta *int64
		Value *float64
	}
	tests := []struct {
		name     string
		data     []entity.Alert
		arg      argument
		want     want
		wantCode codes.Code
		wantErr  bool
	}{
		{
			name: "invalid type given",
			arg: argument{
				ID:   "alert1",
				Type: "invalidType",
			},
			wantCode: codes.InvalidArgument,
			wantErr:  true,
		},
		{
			name: "invalid name argument given",
			arg: argument{
				ID:   "",
				Type: "gauge",
			},
			wantCode: codes.InvalidArgument,
			wantErr:  true,
		},
		{
			name: "success counter case with empty storage",
			arg: argument{
				ID:    "alert1",
				Type:  "counter",
				Delta: intPointer(1),
			},
			want: want{
				ID:    "alert1",
				Type:  "counter",
				Delta: intPointer(1),
			},
			wantErr: false,
		},
		{
			name: "success counter with filled storage",
			data: []entity.Alert{
				{
					IntValue: intPointer(1),
					Type:     "counter",
					Name:     "alert1",
				},
			},
			arg: argument{
				ID:    "alert1",
				Type:  "counter",
				Delta: intPointer(1),
			},
			want: want{
				ID:    "alert1",
				Type:  "counter",
				Delta: intPointer(2),
			},
			wantErr: false,
		},
		{
			name: "success gauge with empty storage",
			arg: argument{
				ID:    "alert1",
				Type:  "gauge",
				Value: floatPointer(1),
			},
			want: want{
				ID:    "alert1",
				Type:  "gauge",
				Value: floatPointer(1),
			},
			wantErr: false,
		},
		{
			name: "success gauge with filled storage",
			data: []entity.Alert{
				{
					FloatValue: floatPointer(1),
					Type:       "gauge",
					Name:       "alert1",
				},
			},
			arg: argument{
				ID:    "alert1",
				Type:  "gauge",
				Value: floatPointer(2),
			},
			want: want{
				ID:    "alert1",
				Type:  "gauge",
				Value: floatPointer(2),
			},
			wantErr: false,
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, d := range tt.data {
				err := strg.Save(ctx, d.Name, d)
				require.NoError(t, err)
			}

			in := &pb.UpdateMetricsRequest{Metrics: &pb.Metrics{
				Value: tt.arg.Value,
				Delta: tt.arg.Delta,
				Id:    tt.arg.ID,
				Type:  tt.arg.Type,
			}}
			resp, err := client.Update(ctx, in)
			if tt.wantErr {
				require.Error(t, err)
				e, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, e.Code())
				return
			} else {
				require.NoError(t, err)
			}

			got := resp.Metrics
			assert.Equal(t, tt.want.ID, got.Id)
			assert.Equal(t, tt.want.Type, got.Type)

			if tt.want.Value == nil {
				assert.Nil(t, got.Value)
			} else {
				if got.Value == nil {
					t.Errorf("value of field `Value` expected to be not nil.")
					return
				}
				assert.Equal(t, *tt.want.Value, *got.Value)
			}

			if tt.want.Delta == nil {
				assert.Nil(t, got.Delta)
			} else {
				if got.Delta == nil {
					t.Errorf("value of field `Delta` expected to be not nil")
					return
				}
				assert.Equal(t, *tt.want.Delta, *got.Delta)
			}

			strg.Reset()
		})
	}
}

func floatPointer(value float64) *float64 {
	return &value
}

func intPointer(value int64) *int64 {
	return &value
}
