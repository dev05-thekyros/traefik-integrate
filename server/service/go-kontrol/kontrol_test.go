package gokontrol

import (
	"context"
	"github.com/golang/mock/gomock"
	"gorm.io/gorm"
	"reflect"
	"testing"
	"time"
)

func TestDefaultKontrol_ValidateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		store  KontrolStore
		Option KontrolOption
	}
	type args struct {
		c         context.Context
		jwtToken  string
		reqPath   string
		reqMethod string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Object
		wantErr bool
	}{
		{name: "#1: Jwt token is invalid",
			fields: fields{
				store:  NewMockKontrolStore(ctrl),
				Option: DefaultKontrolOption,
			},
			args: args{
				c:         context.Background(),
				jwtToken:  "invalid-token",
				reqPath:   "/oauth",
				reqMethod: "POST",
			},
			want:    nil,
			wantErr: true,
		},
		{name: "#2: Token is valid,  verify service follow path and request path not exist",
			fields: fields{
				store: func() KontrolStore {
					kontrolStore := NewMockKontrolStore(ctrl)
					kontrolStore.EXPECT().GetServiceByExternalId(gomock.Any(), "idt").Return(nil, gorm.ErrRecordNotFound).AnyTimes()
					return kontrolStore
				}(),
				Option: DefaultKontrolOption,
			},
			args: args{
				c:         context.Background(),
				jwtToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwZXJtaXNzaW9uIjp7Ijg4NTAzMzk4LWRiMGYtMTFlYy05ZDY0LTAyNDJhYzEyMDAwMiI6eyJlZGl0X3Byb2ZpbGUiOnRydWUsInZpZXdfcHJvZmlsZSI6dHJ1ZX19LCJ0b2tlbiI6InFVenBWSVB6UmNDQjhNcmM1eEJKR0dJeWFoZmdNblJVSlAyUjZYcGxYZWM9IiwiZXhwIjoxNzIxMTM4NjY4fQ.RSk-cmsBznzeum6xUaH3xpJ17r3mE0P5CyooTbAfPP8",
				reqPath:   "/idt/edit-profile",
				reqMethod: "POST",
			},
			want:    nil,
			wantErr: true,
		},
		{name: "#3: Token has valid format. Verify token valid in database and it not valid",
			fields: fields{
				store: func() KontrolStore {
					kontrolStore := NewMockKontrolStore(ctrl)
					kontrolStore.EXPECT().GetServiceByExternalId(gomock.Any(), "dummy-service").Return(&Service{}, nil).AnyTimes()
					kontrolStore.EXPECT().GetObjectByToken(gomock.Any(), "qUzpVIPzRcCB8Mrc5xBJGGIyahfgMnRUJP2R6XplXec=", time.Now().Unix()).Return(nil, gorm.ErrRecordNotFound).AnyTimes()
					return kontrolStore
				}(),
				Option: DefaultKontrolOption,
			},
			args: args{
				c:         context.Background(),
				jwtToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwZXJtaXNzaW9uIjp7ImR1bW15LXNlcnZpY2UtaWQiOnsiZWRpdF9wcm9maWxlIjp0cnVlLCJ2aWV3X3Byb2ZpbGUiOnRydWV9fSwidG9rZW4iOiJxVXpwVklQelJjQ0I4TXJjNXhCSkdHSXlhaGZnTW5SVUpQMlI2WHBsWGVjPSIsImV4cCI6MTcyMTEzODY2OH0.-xpsm3ZqzJy-H4fd3FxQ7HS40CKM3vexdOuUraJrD2o",
				reqPath:   "/dummy-service/edit-profile",
				reqMethod: "POST",
			},
			want:    nil,
			wantErr: true,
		},
		{name: "#4: Verify permission access. Call internal service --> return information of authenticated",
			fields: fields{
				store: func() KontrolStore {
					kontrolStore := NewMockKontrolStore(ctrl)
					kontrolStore.EXPECT().GetServiceByExternalId(gomock.Any(), "dummy-service").Return(&Service{ID: "dummy-service-id"}, nil).AnyTimes()
					kontrolStore.EXPECT().GetObjectByToken(gomock.Any(), "qUzpVIPzRcCB8Mrc5xBJGGIyahfgMnRUJP2R6XplXec=", time.Now().Unix()).Return(&Object{ServiceID: "dummy-service-id"}, nil).AnyTimes()
					return kontrolStore
				}(),
				Option: DefaultKontrolOption,
			},
			args: args{
				c:         context.Background(),
				jwtToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwZXJtaXNzaW9uIjp7ImR1bW15LXNlcnZpY2UtaWQiOnsiZWRpdF9wcm9maWxlIjpmYWxzZSwidmlld19wcm9maWxlIjp0cnVlfX0sInRva2VuIjoicVV6cFZJUHpSY0NCOE1yYzV4QkpHR0l5YWhmZ01uUlVKUDJSNlhwbFhlYz0iLCJleHAiOjE3MjExMzg2Njh9.WI3fgTKGIrBYbdzOsjjL612UVlFTVP0CWGhWqCmbuEc",
				reqPath:   "/dummy-service/edit-profile",
				reqMethod: "POST",
			},
			want:    &Object{ServiceID: "dummy-service-id"},
			wantErr: false,
		},
		{name: "#5: Verify permission access. Call permission of cross services. Check Permission through path --> turned off --> Call internal service --> return information of authenticated",
			fields: fields{
				store: func() KontrolStore {
					kontrolStore := NewMockKontrolStore(ctrl)
					kontrolStore.EXPECT().GetServiceByExternalId(gomock.Any(), "dummy-service").Return(&Service{ID: "dummy-service-id"}, nil).AnyTimes()
					kontrolStore.EXPECT().GetObjectByToken(gomock.Any(), "qUzpVIPzRcCB8Mrc5xBJGGIyahfgMnRUJP2R6XplXec=", time.Now().Unix()).Return(&Object{ServiceID: "another-service-id"}, nil).AnyTimes()
					return kontrolStore
				}(),
				Option: DefaultKontrolOption,
			},
			args: args{
				c:         context.Background(),
				jwtToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwZXJtaXNzaW9uIjp7ImR1bW15LXNlcnZpY2UtaWQiOnsiUE9TVEAvZWRpdC1wcm9maWxlIjpmYWxzZSwidmlld19wcm9maWxlIjp0cnVlfX0sInRva2VuIjoicVV6cFZJUHpSY0NCOE1yYzV4QkpHR0l5YWhmZ01uUlVKUDJSNlhwbFhlYz0iLCJleHAiOjE3MjExMzg2Njh9.e2YDdwt2yxyT-zf78ehe0Ph2Bs1N0HE1nLe_2TdOHv0",
				reqPath:   "/dummy-service/edit-profile",
				reqMethod: "POST",
			},
			want:    nil,
			wantErr: true,
		},
		{name: "#6: Verify permission access. Call permission of cross services. Check Permission through path --> turned on --> Call internal service --> return information of authenticated",
			fields: fields{
				store: func() KontrolStore {
					kontrolStore := NewMockKontrolStore(ctrl)
					kontrolStore.EXPECT().GetServiceByExternalId(gomock.Any(), "dummy-service").Return(&Service{ID: "dummy-service-id"}, nil).AnyTimes()
					kontrolStore.EXPECT().GetObjectByToken(gomock.Any(), "qUzpVIPzRcCB8Mrc5xBJGGIyahfgMnRUJP2R6XplXec=", time.Now().Unix()).Return(&Object{ServiceID: "another-service-id"}, nil).AnyTimes()
					return kontrolStore
				}(),
				Option: DefaultKontrolOption,
			},
			args: args{
				c:         context.Background(),
				jwtToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwZXJtaXNzaW9uIjp7ImR1bW15LXNlcnZpY2UtaWQiOnsiUE9TVEAvZWRpdC1wcm9maWxlIjp0cnVlLCJ2aWV3X3Byb2ZpbGUiOnRydWV9fSwidG9rZW4iOiJxVXpwVklQelJjQ0I4TXJjNXhCSkdHSXlhaGZnTW5SVUpQMlI2WHBsWGVjPSIsImV4cCI6MTcyMTEzODY2OH0.S-Q4XFNklF0N2XDSQruNXO5-qgLDeXrjuY44y22WXMw",
				reqPath:   "/dummy-service/edit-profile",
				reqMethod: "POST",
			},
			want:    &Object{ServiceID: "another-service-id"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := NewBasicKontrol(tt.fields.store)
			got, err := k.ValidateToken(tt.args.c, tt.args.jwtToken, tt.args.reqPath, tt.args.reqMethod)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateToken() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultKontrol_CreateCert(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ControlStore := NewMockKontrolStore(ctrl)

	type fields struct {
		store  KontrolStore
		Option KontrolOption
	}
	type args struct {
		obj              *Object
		policy           []*Policy
		enforce          []*Policy
		extendServiceIds []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *CertForSign
		want1   string
		want2   string
		wantErr bool
	}{
		{name: "#1: generate token from policy, should successfully",
			fields: fields{
				store:  ControlStore,
				Option: DefaultKontrolOption,
			},
			args: args{
				obj: &Object{
					ID:         "hash-obj-1",
					GlobalID:   "generated-uuid-gid",
					ExternalID: "dummy service",
					ServiceID:  "generated-uuid-sid",
					ExpiryDate: 1689594899,
					Attributes: map[string]interface{}{}, // ignore attributes, plan it for extensions feature
					ApplyPolicy: []*Policy{
						&Policy{
							ID:        "generated-uuid-dpid1",
							Name:      "Customize policy for another service",
							ServiceID: "generated-uuid-osid1",
							Permission: map[string]int{
								"view_hr_profile": PolicyPermission.TRUE,
							},
						},
						&Policy{
							ID:        "generated-uuid-dpid1",
							Name:      "On the fly customize policy for dummy service",
							ServiceID: "generated-uuid-sid",
							Permission: map[string]int{
								"update_profile": PolicyPermission.FALSE,
							},
						},
					},
				},
				policy: []*Policy{
					&Policy{
						ID:        "generated-uuid-dpid1",
						Name:      "dummy service default policy 1",
						ServiceID: "generated-uuid-sid",
						Permission: map[string]int{
							"view_profile":   PolicyPermission.TRUE,
							"update_profile": PolicyPermission.TRUE,
							"delete_profile": PolicyPermission.TRUE,
						},
					},
				},
				enforce: []*Policy{
					&Policy{
						ID:        "generated-uuid-epid1",
						Name:      "dummy service enforce policy 1",
						ServiceID: "generated-uuid-sid",
						Permission: map[string]int{
							"delete_profile": PolicyPermission.FALSE, // admin turn off feature delete profile for all policies in this service
						},
					},
				},
				extendServiceIds: []string{},
			},
			want: &CertForSign{
				ID:         "hash-obj-1",
				GlobalID:   "generated-uuid-gid",
				ExternalID: "dummy service",
				ServiceID:  "generated-uuid-sid",
				ExpiryDate: 1689594899,
				Attributes: map[string]interface{}{}, // ignore attributes, plan it for extensions feature
				Permission: map[string]map[string]bool{
					"generated-uuid-sid": {
						"view_profile": true,
					},
					"generated-uuid-osid1": {
						"view_hr_profile": true,
					},
				},
			},
			want1:   "FUkAkUDZLmoYZc4JxJR2oSai7Ivm_nLLJEpxau5hHbc=",
			want2:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwZXJtaXNzaW9uIjp7ImdlbmVyYXRlZC11dWlkLW9zaWQxIjp7InZpZXdfaHJfcHJvZmlsZSI6dHJ1ZX0sImdlbmVyYXRlZC11dWlkLXNpZCI6eyJ2aWV3X3Byb2ZpbGUiOnRydWV9fSwidG9rZW4iOiJGVWtBa1VEWkxtb1laYzRKeEpSMm9TYWk3SXZtX25MTEpFcHhhdTVoSGJjPSIsImV4cCI6MTY4OTU5NDg5OX0.1JUMeq_vCcIr_lUlwQefkyKJDAApmS0V6QgC0bZK7dE",
			wantErr: false,
		},
		{name: "#2: generate token from policy, with extendServiceIds ",
			fields: fields{
				store:  ControlStore,
				Option: DefaultKontrolOption,
			},
			args: args{
				obj: &Object{
					ID:         "hash-obj-1",
					GlobalID:   "generated-uuid-gid",
					ExternalID: "dummy service",
					ServiceID:  "generated-uuid-sid",
					ExpiryDate: 1689594899,
					Attributes: map[string]interface{}{}, // ignore attributes, plan it for extensions feature
					ApplyPolicy: []*Policy{
						&Policy{
							ID:        "generated-uuid-dpid1",
							Name:      "Customize policy for another service",
							ServiceID: "generated-uuid-osid1",
							Permission: map[string]int{
								"view_hr_profile": PolicyPermission.TRUE,
							},
						},
						&Policy{
							ID:        "generated-uuid-dpid1",
							Name:      "On the fly customize policy for dummy service",
							ServiceID: "generated-uuid-sid",
							Permission: map[string]int{
								"update_profile": PolicyPermission.FALSE,
							},
						},
					},
				},
				policy: []*Policy{
					&Policy{
						ID:        "generated-uuid-dpid1",
						Name:      "dummy service default policy 1",
						ServiceID: "generated-uuid-sid",
						Permission: map[string]int{
							"view_profile":   PolicyPermission.TRUE,
							"update_profile": PolicyPermission.TRUE,
							"delete_profile": PolicyPermission.TRUE,
						},
					},
				},
				enforce: []*Policy{
					&Policy{
						ID:        "generated-uuid-epid1",
						Name:      "dummy service enforce policy 1",
						ServiceID: "generated-uuid-sid",
						Permission: map[string]int{
							"delete_profile": PolicyPermission.FALSE, // admin turn off feature delete profile for all policies in this service
						},
					},
				},
				extendServiceIds: []string{"uuid-sap-service-id"},
			},
			want: &CertForSign{
				ID:         "hash-obj-1",
				GlobalID:   "generated-uuid-gid",
				ExternalID: "dummy service",
				ServiceID:  "generated-uuid-sid",
				ExpiryDate: 1689594899,
				Attributes: map[string]interface{}{}, // ignore attributes, plan it for extensions feature
				Permission: map[string]map[string]bool{
					"generated-uuid-sid": {
						"view_profile": true,
					},
					"generated-uuid-osid1": {
						"view_hr_profile": true,
					},
					"uuid-sap-service-id": {},
				},
			},
			want1:   "eTGgDgED0z_Tgf8dNxOLNfv_NPq8pj6c6t8jU-Ob11U=",
			want2:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwZXJtaXNzaW9uIjp7ImdlbmVyYXRlZC11dWlkLW9zaWQxIjp7InZpZXdfaHJfcHJvZmlsZSI6dHJ1ZX0sImdlbmVyYXRlZC11dWlkLXNpZCI6eyJ2aWV3X3Byb2ZpbGUiOnRydWV9LCJ1dWlkLXNhcC1zZXJ2aWNlLWlkIjp7fX0sInRva2VuIjoiZVRHZ0RnRUQwel9UZ2Y4ZE54T0xOZnZfTlBxOHBqNmM2dDhqVS1PYjExVT0iLCJleHAiOjE2ODk1OTQ4OTl9.NlqcKBmHbOli19cvJuFN8vGfF86VHJPcN5Tr7RRc0bk",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := DefaultKontrol{
				store:  tt.fields.store,
				Option: tt.fields.Option,
			}
			got, got1, got2, err := k.CreateCert(tt.args.obj, tt.args.policy, tt.args.enforce, tt.args.extendServiceIds)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateCert() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("CreateCert() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("CreateCert() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
