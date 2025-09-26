package client

import (
	"context"
	"reflect"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"go.uber.org/mock/gomock"
	"k8s.io/utils/ptr"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client/mock_client"
)

const (
	lbID       = "11111111-1111-1111-1111-111111111111"
	lbIPID     = "11111111-1111-1111-1111-111111111111"
	backendID  = "11111111-1111-1111-1111-111111111111"
	frontendID = "11111111-1111-1111-1111-111111111111"
	aclID      = "11111111-1111-1111-1111-111111111111"
)

var (
	lb1 = lb.LB{
		ID:   lbID,
		Name: "lb-name-1",
		Tags: []string{"tag1", "tag2"},
	}
	lb2 = lb.LB{
		ID:   "22222222-2222-2222-2222-22222222222",
		Name: "lb-name-2",
		Tags: []string{"tag1", "tag2"},
	}
)

func TestClient_FindLB(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx  context.Context
		zone scw.Zone
		tags []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *lb.LB
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "no lb found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				tags: []string{"tag1", "tag2"},
			},
			wantErr: true,
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.ListLBs(&lb.ZonedAPIListLBsRequest{
					Zone:      scw.ZoneFrPar1,
					Tags:      []string{"tag1", "tag2"},
					ProjectID: ptr.To(projectID),
				}, gomock.Any(), gomock.Any()).Return(&lb.ListLBsResponse{
					LBs: []*lb.LB{},
				}, nil)
			},
		},
		{
			name: "one lb found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				tags: []string{"tag1", "tag2"},
			},
			want: &lb.LB{
				ID:   lbID,
				Name: "lb-name",
				Tags: []string{"tag1", "tag2"},
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.ListLBs(&lb.ZonedAPIListLBsRequest{
					Zone:      scw.ZoneFrPar1,
					Tags:      []string{"tag1", "tag2"},
					ProjectID: ptr.To(projectID),
				}, gomock.Any(), gomock.Any()).Return(&lb.ListLBsResponse{
					LBs: []*lb.LB{
						{
							ID:   lbID,
							Name: "lb-name",
							Tags: []string{"tag1", "tag2"},
						},
					},
				}, nil)
			},
		},
		{
			name: "multiple lbs found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				tags: []string{"tag1", "tag2"},
			},
			wantErr: true,
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.ListLBs(&lb.ZonedAPIListLBsRequest{
					Zone:      scw.ZoneFrPar1,
					Tags:      []string{"tag1", "tag2"},
					ProjectID: ptr.To(projectID),
				}, gomock.Any(), gomock.Any()).Return(&lb.ListLBsResponse{
					LBs: []*lb.LB{
						{
							Name: "lb-name-1",
							Tags: []string{"tag1", "tag2"},
						},
						{
							Name: "lb-name-2",
							Tags: []string{"tag1", "tag2"},
						},
					},
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			lbMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			got, err := c.FindLB(tt.args.ctx, tt.args.zone, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindLB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindLB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_MigrateLB(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx     context.Context
		zone    scw.Zone
		id      string
		newType string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *lb.LB
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "migrate lb to new type",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:     context.TODO(),
				zone:    scw.ZoneFrPar1,
				id:      lbID,
				newType: "LB-GP-M",
			},
			want: &lb.LB{
				ID:   lbID,
				Name: "lb-name",
				Type: "lb-gp-m",
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.MigrateLB(&lb.ZonedAPIMigrateLBRequest{
					Zone: scw.ZoneFrPar1,
					LBID: lbID,
					Type: "lb-gp-m",
				}, gomock.Any()).Return(&lb.LB{
					ID:   lbID,
					Name: "lb-name",
					Type: "lb-gp-m",
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			lbMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			got, err := c.MigrateLB(tt.args.ctx, tt.args.zone, tt.args.id, tt.args.newType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.MigrateLB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.MigrateLB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_FindLBIP(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx  context.Context
		zone scw.Zone
		ip   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *lb.IP
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "no lb ip found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				ip:   "42.42.42.42",
			},
			wantErr: true,
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.ListIPs(&lb.ZonedAPIListIPsRequest{
					Zone:      scw.ZoneFrPar1,
					ProjectID: ptr.To(projectID),
					IPAddress: ptr.To("42.42.42.42"),
				}, gomock.Any()).Return(&lb.ListIPsResponse{
					IPs: []*lb.IP{},
				}, nil)
			},
		},
		{
			name: "lb ip found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				ip:   "42.42.42.42",
			},
			want: &lb.IP{
				IPAddress: "42.42.42.42",
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.ListIPs(&lb.ZonedAPIListIPsRequest{
					Zone:      scw.ZoneFrPar1,
					ProjectID: ptr.To(projectID),
					IPAddress: ptr.To("42.42.42.42"),
				}, gomock.Any()).Return(&lb.ListIPsResponse{
					IPs:        []*lb.IP{{IPAddress: "42.42.42.42"}},
					TotalCount: 1,
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			lbMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			got, err := c.FindLBIP(tt.args.ctx, tt.args.zone, tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindLBIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindLBIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_CreateLB(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx     context.Context
		zone    scw.Zone
		name    string
		lbType  string
		ipID    *string
		private bool
		tags    []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *lb.LB
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "create lb with ip",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:    context.TODO(),
				zone:   scw.ZoneFrPar1,
				name:   "my-lb",
				lbType: "LB-GP-M",
				ipID:   ptr.To(lbIPID),
				tags:   []string{"tag1", "tag2"},
			},
			want: &lb.LB{
				ID:   lbID,
				Name: "my-lb",
				Type: "lb-gp-m",
				IP:   []*lb.IP{{ID: lbIPID, IPAddress: "42.42.42.42"}},
				Tags: []string{"tag1", "tag2"},
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.CreateLB(&lb.ZonedAPICreateLBRequest{
					Zone:               scw.ZoneFrPar1,
					Name:               "my-lb",
					Type:               "lb-gp-m",
					IPIDs:              []string{lbIPID},
					Tags:               []string{"tag1", "tag2", createdByTag},
					Description:        createdByDescription,
					AssignFlexibleIP:   ptr.To(false),
					AssignFlexibleIPv6: ptr.To(false),
				}, gomock.Any()).Return(&lb.LB{
					ID:   lbID,
					Name: "my-lb",
					Type: "lb-gp-m",
					IP:   []*lb.IP{{ID: lbIPID, IPAddress: "42.42.42.42"}},
					Tags: []string{"tag1", "tag2"},
				}, nil)
			},
		},
		{
			name: "create lb without ip",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:    context.TODO(),
				zone:   scw.ZoneFrPar1,
				name:   "my-lb",
				lbType: "LB-GP-M",
				tags:   []string{"tag1", "tag2"},
			},
			want: &lb.LB{
				ID:   lbID,
				Name: "my-lb",
				Type: "lb-gp-m",
				IP:   []*lb.IP{{ID: lbIPID, IPAddress: "42.42.42.42"}},
				Tags: []string{"tag1", "tag2"},
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.CreateLB(&lb.ZonedAPICreateLBRequest{
					Zone:               scw.ZoneFrPar1,
					Name:               "my-lb",
					Type:               "lb-gp-m",
					Tags:               []string{"tag1", "tag2", createdByTag},
					Description:        createdByDescription,
					AssignFlexibleIP:   ptr.To(true),
					AssignFlexibleIPv6: ptr.To(false),
				}, gomock.Any()).Return(&lb.LB{
					ID:   lbID,
					Name: "my-lb",
					Type: "lb-gp-m",
					IP:   []*lb.IP{{ID: lbIPID, IPAddress: "42.42.42.42"}},
					Tags: []string{"tag1", "tag2"},
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			lbMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			got, err := c.CreateLB(tt.args.ctx, tt.args.zone, tt.args.name, tt.args.lbType, tt.args.ipID, tt.args.private, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.CreateLB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.CreateLB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_DeleteLB(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx       context.Context
		zone      scw.Zone
		id        string
		releaseIP bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "delete lb with release IP",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:       context.TODO(),
				zone:      scw.ZoneFrPar1,
				id:        lbID,
				releaseIP: true,
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.DeleteLB(&lb.ZonedAPIDeleteLBRequest{
					Zone:      scw.ZoneFrPar1,
					LBID:      lbID,
					ReleaseIP: true,
				}, gomock.Any()).Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			lbMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			if err := c.DeleteLB(tt.args.ctx, tt.args.zone, tt.args.id, tt.args.releaseIP); (err != nil) != tt.wantErr {
				t.Errorf("Client.DeleteLB() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_FindLBs(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx  context.Context
		tags []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*lb.LB
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "no lbs found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				tags: []string{"tag1", "tag2"},
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.Zones()
				l.ListLBs(&lb.ZonedAPIListLBsRequest{
					Zone:      scw.ZoneFrPar1,
					Tags:      []string{"tag1", "tag2"},
					ProjectID: ptr.To(projectID),
				}, gomock.Any(), gomock.Any(), gomock.Any()).Return(&lb.ListLBsResponse{
					LBs: []*lb.LB{},
				}, nil)
			},
			want: []*lb.LB{},
		},
		{
			name: "multiple lbs found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				tags: []string{"tag1", "tag2"},
			},
			want: []*lb.LB{&lb1, &lb2},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.Zones()
				l.ListLBs(&lb.ZonedAPIListLBsRequest{
					Zone:      scw.ZoneFrPar1,
					Tags:      []string{"tag1", "tag2"},
					ProjectID: ptr.To(projectID),
				}, gomock.Any(), gomock.Any(), gomock.Any()).Return(&lb.ListLBsResponse{
					TotalCount: 2,
					LBs:        []*lb.LB{&lb1, &lb2},
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			got, err := c.FindLBs(tt.args.ctx, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindLBs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindLBs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_FindBackend(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx  context.Context
		zone scw.Zone
		lbID string
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *lb.Backend
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "no backend found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				lbID: lbID,
				name: "backend-name",
			},
			wantErr: true,
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.ListBackends(&lb.ZonedAPIListBackendsRequest{
					Zone: scw.ZoneFrPar1,
					LBID: lbID,
					Name: ptr.To("backend-name"),
				}, gomock.Any(), gomock.Any()).Return(&lb.ListBackendsResponse{
					Backends: []*lb.Backend{},
				}, nil)
			},
		},
		{
			name: "backend found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				lbID: lbID,
				name: "backend-name",
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.ListBackends(&lb.ZonedAPIListBackendsRequest{
					Zone: scw.ZoneFrPar1,
					LBID: lbID,
					Name: ptr.To("backend-name"),
				}, gomock.Any(), gomock.Any()).Return(&lb.ListBackendsResponse{
					Backends: []*lb.Backend{
						{
							ID:   backendID,
							Name: "backend-name",
						},
					},
				}, nil)
			},
			want: &lb.Backend{
				ID:   backendID,
				Name: "backend-name",
			},
		},
		{
			name: "duplicate backend found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				lbID: lbID,
				name: "backend-name",
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.ListBackends(&lb.ZonedAPIListBackendsRequest{
					Zone: scw.ZoneFrPar1,
					LBID: lbID,
					Name: ptr.To("backend-name"),
				}, gomock.Any(), gomock.Any()).Return(&lb.ListBackendsResponse{
					Backends: []*lb.Backend{
						{Name: "backend-name"},
						{Name: "backend-name"},
					},
				}, nil)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			lbMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			got, err := c.FindBackend(tt.args.ctx, tt.args.zone, tt.args.lbID, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindBackend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindBackend() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_CreateBackend(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx     context.Context
		zone    scw.Zone
		lbID    string
		name    string
		servers []string
		port    int32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *lb.Backend
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "create backend",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:     context.TODO(),
				zone:    scw.ZoneFrPar1,
				lbID:    lbID,
				name:    "backend-name",
				servers: []string{"42.42.42.42"},
				port:    6443,
			},
			want: &lb.Backend{
				ID:   backendID,
				Name: "backend-name",
				Pool: []string{"42.42.42.42"},
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.CreateBackend(&lb.ZonedAPICreateBackendRequest{
					Zone:            scw.ZoneFrPar1,
					LBID:            lbID,
					Name:            "backend-name",
					ForwardProtocol: lb.ProtocolTCP,
					ForwardPort:     6443,
					HealthCheck: &lb.HealthCheck{
						Port:            6443,
						CheckMaxRetries: 5,
						TCPConfig:       &lb.HealthCheckTCPConfig{},
					},
					ServerIP: []string{"42.42.42.42"},
				}, gomock.Any()).Return(&lb.Backend{
					ID:   backendID,
					Name: "backend-name",
					Pool: []string{"42.42.42.42"},
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			lbMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			got, err := c.CreateBackend(tt.args.ctx, tt.args.zone, tt.args.lbID, tt.args.name, tt.args.servers, tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.CreateBackend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.CreateBackend() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_SetBackendServers(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx       context.Context
		zone      scw.Zone
		backendID string
		servers   []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *lb.Backend
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "set backend servers",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:       context.TODO(),
				zone:      scw.ZoneFrPar1,
				backendID: backendID,
				servers:   []string{"42.42.42.42"},
			},
			want: &lb.Backend{
				ID:   backendID,
				Name: "backend-name",
				Pool: []string{"42.42.42.42"},
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.SetBackendServers(&lb.ZonedAPISetBackendServersRequest{
					Zone:      scw.ZoneFrPar1,
					BackendID: backendID,
					ServerIP:  []string{"42.42.42.42"},
				}, gomock.Any()).Return(&lb.Backend{
					ID:   backendID,
					Name: "backend-name",
					Pool: []string{"42.42.42.42"},
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			lbMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			got, err := c.SetBackendServers(tt.args.ctx, tt.args.zone, tt.args.backendID, tt.args.servers)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.SetBackendServers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.SetBackendServers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_FindFrontend(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx  context.Context
		zone scw.Zone
		lbID string
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *lb.Frontend
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "no frontend found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				lbID: lbID,
				name: "frontend-name",
			},
			wantErr: true,
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.ListFrontends(&lb.ZonedAPIListFrontendsRequest{
					Zone: scw.ZoneFrPar1,
					LBID: lbID,
					Name: ptr.To("frontend-name"),
				}, gomock.Any(), gomock.Any()).Return(&lb.ListFrontendsResponse{
					Frontends: []*lb.Frontend{},
				}, nil)
			},
		},
		{
			name: "frontend found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				lbID: lbID,
				name: "frontend-name",
			},
			want: &lb.Frontend{
				ID:   frontendID,
				Name: "frontend-name",
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.ListFrontends(&lb.ZonedAPIListFrontendsRequest{
					Zone: scw.ZoneFrPar1,
					LBID: lbID,
					Name: ptr.To("frontend-name"),
				}, gomock.Any(), gomock.Any()).Return(&lb.ListFrontendsResponse{
					Frontends: []*lb.Frontend{
						{
							ID:   frontendID,
							Name: "frontend-name",
						},
					},
				}, nil)
			},
		},
		{
			name: "duplicate frontend found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				lbID: lbID,
				name: "frontend-name",
			},
			wantErr: true,
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.ListFrontends(&lb.ZonedAPIListFrontendsRequest{
					Zone: scw.ZoneFrPar1,
					LBID: lbID,
					Name: ptr.To("frontend-name"),
				}, gomock.Any(), gomock.Any()).Return(&lb.ListFrontendsResponse{
					Frontends: []*lb.Frontend{
						{Name: "frontend-name"},
						{Name: "frontend-name"},
					},
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			lbMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			got, err := c.FindFrontend(tt.args.ctx, tt.args.zone, tt.args.lbID, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindFrontend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindFrontend() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_CreateFrontend(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx       context.Context
		zone      scw.Zone
		lbID      string
		name      string
		backendID string
		port      int32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *lb.Frontend
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "create frontend",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:       context.TODO(),
				zone:      scw.ZoneFrPar1,
				lbID:      lbID,
				name:      "frontend-name",
				backendID: backendID,
				port:      443,
			},
			want: &lb.Frontend{
				ID:          frontendID,
				Name:        "frontend-name",
				InboundPort: 443,
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.CreateFrontend(&lb.ZonedAPICreateFrontendRequest{
					Zone:        scw.ZoneFrPar1,
					LBID:        lbID,
					Name:        "frontend-name",
					BackendID:   backendID,
					InboundPort: 443,
				}, gomock.Any()).Return(&lb.Frontend{
					ID:          frontendID,
					Name:        "frontend-name",
					InboundPort: 443,
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			lbMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			got, err := c.CreateFrontend(tt.args.ctx, tt.args.zone, tt.args.lbID, tt.args.name, tt.args.backendID, tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.CreateFrontend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.CreateFrontend() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_FindLBPrivateNetwork(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx              context.Context
		zone             scw.Zone
		lbID             string
		privateNetworkID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *lb.PrivateNetwork
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "no private network found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:              context.TODO(),
				zone:             scw.ZoneFrPar1,
				lbID:             lbID,
				privateNetworkID: privateNetworkID,
			},
			wantErr: true,
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.ListLBPrivateNetworks(&lb.ZonedAPIListLBPrivateNetworksRequest{
					Zone: scw.ZoneFrPar1,
					LBID: lbID,
				}, gomock.Any(), gomock.Any()).Return(&lb.ListLBPrivateNetworksResponse{
					PrivateNetwork: []*lb.PrivateNetwork{},
				}, nil)
			},
		},
		{
			name: "private network found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:              context.TODO(),
				zone:             scw.ZoneFrPar1,
				lbID:             lbID,
				privateNetworkID: privateNetworkID,
			},
			want: &lb.PrivateNetwork{
				LB:               &lb1,
				PrivateNetworkID: privateNetworkID,
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.ListLBPrivateNetworks(&lb.ZonedAPIListLBPrivateNetworksRequest{
					Zone: scw.ZoneFrPar1,
					LBID: lbID,
				}, gomock.Any(), gomock.Any()).Return(&lb.ListLBPrivateNetworksResponse{
					PrivateNetwork: []*lb.PrivateNetwork{
						{LB: &lb1, PrivateNetworkID: privateNetworkID},
					},
				}, nil)
			},
		},
		{
			name: "duplicate private network found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:              context.TODO(),
				zone:             scw.ZoneFrPar1,
				lbID:             lbID,
				privateNetworkID: privateNetworkID,
			},
			wantErr: true,
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.ListLBPrivateNetworks(&lb.ZonedAPIListLBPrivateNetworksRequest{
					Zone: scw.ZoneFrPar1,
					LBID: lbID,
				}, gomock.Any(), gomock.Any()).Return(&lb.ListLBPrivateNetworksResponse{
					PrivateNetwork: []*lb.PrivateNetwork{
						{LB: &lb1, PrivateNetworkID: privateNetworkID},
						{LB: &lb1, PrivateNetworkID: privateNetworkID},
					},
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			lbMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			got, err := c.FindLBPrivateNetwork(tt.args.ctx, tt.args.zone, tt.args.lbID, tt.args.privateNetworkID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindLBPrivateNetwork() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindLBPrivateNetwork() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_AttachLBPrivateNetwork(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx              context.Context
		zone             scw.Zone
		lbID             string
		privateNetworkID string
		ipID             *string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "attach private network to lb",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:              context.TODO(),
				zone:             scw.ZoneFrPar1,
				lbID:             lbID,
				privateNetworkID: privateNetworkID,
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.AttachPrivateNetwork(&lb.ZonedAPIAttachPrivateNetworkRequest{
					Zone:             scw.ZoneFrPar1,
					LBID:             lbID,
					PrivateNetworkID: privateNetworkID,
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			lbMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			if err := c.AttachLBPrivateNetwork(tt.args.ctx, tt.args.zone, tt.args.lbID, tt.args.privateNetworkID, tt.args.ipID); (err != nil) != tt.wantErr {
				t.Errorf("Client.AttachLBPrivateNetwork() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_ListLBACLs(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx        context.Context
		zone       scw.Zone
		frontendID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*lb.ACL
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "list acls for frontend",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:        context.TODO(),
				zone:       scw.ZoneFrPar1,
				frontendID: frontendID,
			},
			want: []*lb.ACL{{ID: aclID, Name: "acl-name"}},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.ListACLs(&lb.ZonedAPIListACLsRequest{
					Zone:       scw.ZoneFrPar1,
					FrontendID: frontendID,
				}, gomock.Any(), gomock.Any()).Return(&lb.ListACLResponse{
					ACLs: []*lb.ACL{{ID: aclID, Name: "acl-name"}},
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			lbMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			got, err := c.ListLBACLs(tt.args.ctx, tt.args.zone, tt.args.frontendID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.ListLBACLs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.ListLBACLs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_SetLBACLs(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx        context.Context
		zone       scw.Zone
		frontendID string
		acls       []*lb.ACLSpec
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "set acls for frontend",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:        context.TODO(),
				zone:       scw.ZoneFrPar1,
				frontendID: frontendID,
				acls:       []*lb.ACLSpec{{Name: "acl-name"}},
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.SetACLs(&lb.ZonedAPISetACLsRequest{
					Zone:       scw.ZoneFrPar1,
					FrontendID: frontendID,
					ACLs:       []*lb.ACLSpec{{Name: "acl-name"}},
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			lbMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			if err := c.SetLBACLs(tt.args.ctx, tt.args.zone, tt.args.frontendID, tt.args.acls); (err != nil) != tt.wantErr {
				t.Errorf("Client.SetLBACLs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_FindLBACLByName(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx        context.Context
		zone       scw.Zone
		frontendID string
		name       string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *lb.ACL
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "no acl found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:        context.TODO(),
				zone:       scw.ZoneFrPar1,
				frontendID: frontendID,
				name:       "acl-name",
			},
			wantErr: true,
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.ListACLs(&lb.ZonedAPIListACLsRequest{
					Zone:       scw.ZoneFrPar1,
					FrontendID: frontendID,
					Name:       ptr.To("acl-name"),
				}, gomock.Any(), gomock.Any()).Return(&lb.ListACLResponse{
					ACLs: []*lb.ACL{},
				}, nil)
			},
		},
		{
			name: "acl found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:        context.TODO(),
				zone:       scw.ZoneFrPar1,
				frontendID: frontendID,
				name:       "acl-name",
			},
			want: &lb.ACL{
				ID:   aclID,
				Name: "acl-name",
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.ListACLs(&lb.ZonedAPIListACLsRequest{
					Zone:       scw.ZoneFrPar1,
					FrontendID: frontendID,
					Name:       ptr.To("acl-name"),
				}, gomock.Any(), gomock.Any()).Return(&lb.ListACLResponse{
					TotalCount: 1,
					ACLs:       []*lb.ACL{{ID: aclID, Name: "acl-name"}},
				}, nil)
			},
		},
		{
			name: "duplicate acl found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:        context.TODO(),
				zone:       scw.ZoneFrPar1,
				frontendID: frontendID,
				name:       "acl-name",
			},
			wantErr: true,
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.ListACLs(&lb.ZonedAPIListACLsRequest{
					Zone:       scw.ZoneFrPar1,
					FrontendID: frontendID,
					Name:       ptr.To("acl-name"),
				}, gomock.Any(), gomock.Any()).Return(&lb.ListACLResponse{
					TotalCount: 1,
					ACLs: []*lb.ACL{
						{Name: "acl-name"},
						{Name: "acl-name"},
					},
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			lbMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			got, err := c.FindLBACLByName(tt.args.ctx, tt.args.zone, tt.args.frontendID, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindLBACLByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindLBACLByName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_DeleteLBACL(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx   context.Context
		zone  scw.Zone
		aclID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "delete acl",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:   context.TODO(),
				zone:  scw.ZoneFrPar1,
				aclID: aclID,
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.DeleteACL(&lb.ZonedAPIDeleteACLRequest{
					Zone:  scw.ZoneFrPar1,
					ACLID: aclID,
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			lbMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			if err := c.DeleteLBACL(tt.args.ctx, tt.args.zone, tt.args.aclID); (err != nil) != tt.wantErr {
				t.Errorf("Client.DeleteLBACL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_CreateLBACL(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx            context.Context
		zone           scw.Zone
		frontendID     string
		name           string
		index          int32
		action         lb.ACLActionType
		matchedSubnets []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "create acl",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:            context.TODO(),
				zone:           scw.ZoneFrPar1,
				frontendID:     frontendID,
				name:           "acl-name",
				index:          0,
				action:         lb.ACLActionTypeAllow,
				matchedSubnets: []string{"42.42.42.42/16"},
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.CreateACL(&lb.ZonedAPICreateACLRequest{
					Zone:       scw.ZoneFrPar1,
					FrontendID: frontendID,
					Name:       "acl-name",
					Index:      0,
					Action:     &lb.ACLAction{Type: lb.ACLActionTypeAllow},
					Match:      &lb.ACLMatch{IPSubnet: scw.StringSlicePtr([]string{"42.42.42.42/16"})},
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			lbMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			if err := c.CreateLBACL(tt.args.ctx, tt.args.zone, tt.args.frontendID, tt.args.name, tt.args.index, tt.args.action, tt.args.matchedSubnets); (err != nil) != tt.wantErr {
				t.Errorf("Client.CreateLBACL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_UpdateLBACL(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx            context.Context
		zone           scw.Zone
		aclID          string
		name           string
		index          int32
		action         lb.ACLActionType
		matchedSubnets []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "update acl",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:            context.TODO(),
				zone:           scw.ZoneFrPar1,
				aclID:          aclID,
				name:           "acl",
				index:          0,
				action:         lb.ACLActionTypeAllow,
				matchedSubnets: []string{"192.168.1.0/24"},
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.UpdateACL(&lb.ZonedAPIUpdateACLRequest{
					ACLID:  aclID,
					Zone:   scw.ZoneFrPar1,
					Name:   "acl",
					Index:  0,
					Action: &lb.ACLAction{Type: lb.ACLActionTypeAllow},
					Match:  &lb.ACLMatch{IPSubnet: scw.StringSlicePtr([]string{"192.168.1.0/24"})},
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			lbMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			if err := c.UpdateLBACL(tt.args.ctx, tt.args.zone, tt.args.aclID, tt.args.name, tt.args.index, tt.args.action, tt.args.matchedSubnets); (err != nil) != tt.wantErr {
				t.Errorf("Client.UpdateLBACL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_RemoveBackendServer(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx       context.Context
		zone      scw.Zone
		backendID string
		ip        string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "remove backend server",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:       context.TODO(),
				zone:      scw.ZoneFrPar1,
				backendID: backendID,
				ip:        "42.42.42.42",
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.RemoveBackendServers(&lb.ZonedAPIRemoveBackendServersRequest{
					Zone:      scw.ZoneFrPar1,
					BackendID: backendID,
					ServerIP:  []string{"42.42.42.42"},
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			lbMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			if err := c.RemoveBackendServer(tt.args.ctx, tt.args.zone, tt.args.backendID, tt.args.ip); (err != nil) != tt.wantErr {
				t.Errorf("Client.RemoveBackendServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_AddBackendServer(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx       context.Context
		zone      scw.Zone
		backendID string
		ip        string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(l *mock_client.MockLBAPIMockRecorder)
	}{
		{
			name: "add backend server",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:       context.TODO(),
				zone:      scw.ZoneFrPar1,
				backendID: backendID,
				ip:        "42.42.42.42",
			},
			expect: func(l *mock_client.MockLBAPIMockRecorder) {
				l.AddBackendServers(&lb.ZonedAPIAddBackendServersRequest{
					Zone:      scw.ZoneFrPar1,
					BackendID: backendID,
					ServerIP:  []string{"42.42.42.42"},
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			lbMock := mock_client.NewMockLBAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			lbMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(lbMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				lb:        lbMock,
			}
			if err := c.AddBackendServer(tt.args.ctx, tt.args.zone, tt.args.backendID, tt.args.ip); (err != nil) != tt.wantErr {
				t.Errorf("Client.AddBackendServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
