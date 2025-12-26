package helius

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAsset(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}

			var req map[string]string
			json.NewDecoder(r.Body).Decode(&req)
			if req["id"] != "asset-mint-address" {
				t.Errorf("unexpected id: %s", req["id"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Asset{
				ID:        "asset-mint-address",
				Interface: "V1_NFT",
				Mutable:   true,
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		asset, err := client.GetAsset(context.Background(), "asset-mint-address")

		if err != nil {
			t.Fatalf("GetAsset returned error: %v", err)
		}
		if asset.ID != "asset-mint-address" {
			t.Errorf("ID = %s, want asset-mint-address", asset.ID)
		}
		if asset.Interface != "V1_NFT" {
			t.Errorf("Interface = %s, want V1_NFT", asset.Interface)
		}
	})

	t.Run("empty id", func(t *testing.T) {
		client, _ := NewClient("test-key")
		_, err := client.GetAsset(context.Background(), "")
		if err == nil {
			t.Error("GetAsset should return error for empty id")
		}
	})

	t.Run("asset with content", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Asset{
				ID: "nft-123",
				Content: &AssetContent{
					JSONUri: "https://arweave.net/metadata.json",
					Files: []AssetFile{
						{URI: "https://arweave.net/image.png", Mime: "image/png"},
					},
				},
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		asset, err := client.GetAsset(context.Background(), "nft-123")

		if err != nil {
			t.Fatalf("GetAsset returned error: %v", err)
		}
		if asset.Content == nil {
			t.Fatal("Content should not be nil")
		}
		if asset.Content.JSONUri != "https://arweave.net/metadata.json" {
			t.Errorf("JSONUri = %s, unexpected value", asset.Content.JSONUri)
		}
		if len(asset.Content.Files) != 1 {
			t.Fatalf("len(Files) = %d, want 1", len(asset.Content.Files))
		}
	})

	t.Run("compressed nft", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Asset{
				ID: "cnft-456",
				Compression: &Compression{
					Compressed: true,
					Tree:       "tree-address",
					LeafID:     42,
				},
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		asset, err := client.GetAsset(context.Background(), "cnft-456")

		if err != nil {
			t.Fatalf("GetAsset returned error: %v", err)
		}
		if asset.Compression == nil {
			t.Fatal("Compression should not be nil")
		}
		if !asset.Compression.Compressed {
			t.Error("Compressed should be true")
		}
		if asset.Compression.LeafID != 42 {
			t.Errorf("LeafID = %d, want 42", asset.Compression.LeafID)
		}
	})
}

func TestGetAssetsByOwner(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)
			if req["ownerAddress"] != "owner-wallet" {
				t.Errorf("unexpected ownerAddress: %s", req["ownerAddress"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(AssetsPage{
				Total: 100,
				Limit: 10,
				Items: []Asset{
					{ID: "asset-1"},
					{ID: "asset-2"},
				},
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		page, err := client.GetAssetsByOwner(context.Background(), "owner-wallet", nil)

		if err != nil {
			t.Fatalf("GetAssetsByOwner returned error: %v", err)
		}
		if page.Total != 100 {
			t.Errorf("Total = %d, want 100", page.Total)
		}
		if len(page.Items) != 2 {
			t.Errorf("len(Items) = %d, want 2", len(page.Items))
		}
	})

	t.Run("empty owner address", func(t *testing.T) {
		client, _ := NewClient("test-key")
		_, err := client.GetAssetsByOwner(context.Background(), "", nil)
		if err == nil {
			t.Error("GetAssetsByOwner should return error for empty owner address")
		}
	})

	t.Run("with pagination options", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			if req["page"] != float64(2) {
				t.Errorf("page = %v, want 2", req["page"])
			}
			if req["limit"] != float64(50) {
				t.Errorf("limit = %v, want 50", req["limit"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(AssetsPage{
				Total: 100,
				Limit: 50,
				Page:  2,
				Items: []Asset{},
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		page, err := client.GetAssetsByOwner(context.Background(), "owner-wallet", &AssetsByOwnerOptions{
			Page:  2,
			Limit: 50,
		})

		if err != nil {
			t.Fatalf("GetAssetsByOwner returned error: %v", err)
		}
		if page.Page != 2 {
			t.Errorf("Page = %d, want 2", page.Page)
		}
	})

	t.Run("with display options", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			displayOpts := req["displayOptions"].(map[string]interface{})
			if displayOpts["showFungible"] != true {
				t.Errorf("showFungible = %v, want true", displayOpts["showFungible"])
			}
			if displayOpts["showNativeBalance"] != true {
				t.Errorf("showNativeBalance = %v, want true", displayOpts["showNativeBalance"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(AssetsPage{
				Total:         10,
				Items:         []Asset{},
				NativeBalance: &Balance{Lamports: 1000000000},
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		page, err := client.GetAssetsByOwner(context.Background(), "owner-wallet", &AssetsByOwnerOptions{
			ShowFungible:      true,
			ShowNativeBalance: true,
		})

		if err != nil {
			t.Fatalf("GetAssetsByOwner returned error: %v", err)
		}
		if page.NativeBalance == nil {
			t.Fatal("NativeBalance should not be nil")
		}
		if page.NativeBalance.Lamports != 1000000000 {
			t.Errorf("Lamports = %d, want 1000000000", page.NativeBalance.Lamports)
		}
	})

	t.Run("with cursor pagination", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			if req["cursor"] != "next-page-cursor" {
				t.Errorf("cursor = %v, want next-page-cursor", req["cursor"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(AssetsPage{
				Total:  100,
				Items:  []Asset{},
				Cursor: "another-cursor",
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		page, err := client.GetAssetsByOwner(context.Background(), "owner-wallet", &AssetsByOwnerOptions{
			Cursor: "next-page-cursor",
		})

		if err != nil {
			t.Fatalf("GetAssetsByOwner returned error: %v", err)
		}
		if page.Cursor != "another-cursor" {
			t.Errorf("Cursor = %s, want another-cursor", page.Cursor)
		}
	})
}

func TestSearchAssets(t *testing.T) {
	t.Run("search by owner", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/assets/search" {
				t.Errorf("expected /assets/search, got %s", r.URL.Path)
			}

			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)
			if req["ownerAddress"] != "search-owner" {
				t.Errorf("ownerAddress = %v, want search-owner", req["ownerAddress"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(AssetsPage{
				Total: 5,
				Items: []Asset{{ID: "found-asset"}},
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		page, err := client.SearchAssets(context.Background(), &SearchAssetsOptions{
			OwnerAddress: "search-owner",
		})

		if err != nil {
			t.Fatalf("SearchAssets returned error: %v", err)
		}
		if page.Total != 5 {
			t.Errorf("Total = %d, want 5", page.Total)
		}
	})

	t.Run("nil options", func(t *testing.T) {
		client, _ := NewClient("test-key")
		_, err := client.SearchAssets(context.Background(), nil)
		if err == nil {
			t.Error("SearchAssets should return error for nil options")
		}
	})

	t.Run("search by collection", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)
			if req["groupKey"] != "collection" {
				t.Errorf("groupKey = %v, want collection", req["groupKey"])
			}
			if req["groupValue"] != "collection-address" {
				t.Errorf("groupValue = %v, want collection-address", req["groupValue"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(AssetsPage{Total: 10, Items: []Asset{}})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		_, err := client.SearchAssets(context.Background(), &SearchAssetsOptions{
			GroupKey:   "collection",
			GroupValue: "collection-address",
		})

		if err != nil {
			t.Fatalf("SearchAssets returned error: %v", err)
		}
	})

	t.Run("search compressed only", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)
			if req["compressed"] != true {
				t.Errorf("compressed = %v, want true", req["compressed"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(AssetsPage{Total: 3, Items: []Asset{}})
		}))
		defer server.Close()

		compressed := true
		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		_, err := client.SearchAssets(context.Background(), &SearchAssetsOptions{
			Compressed: &compressed,
		})

		if err != nil {
			t.Fatalf("SearchAssets returned error: %v", err)
		}
	})
}

func TestGetAssetBatch(t *testing.T) {
	t.Run("successful batch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/assets/batch" {
				t.Errorf("expected /assets/batch, got %s", r.URL.Path)
			}

			var req map[string][]string
			json.NewDecoder(r.Body).Decode(&req)
			if len(req["ids"]) != 3 {
				t.Errorf("len(ids) = %d, want 3", len(req["ids"]))
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]Asset{
				{ID: "asset-1"},
				{ID: "asset-2"},
				{ID: "asset-3"},
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		assets, err := client.GetAssetBatch(context.Background(), []string{"asset-1", "asset-2", "asset-3"})

		if err != nil {
			t.Fatalf("GetAssetBatch returned error: %v", err)
		}
		if len(assets) != 3 {
			t.Errorf("len(assets) = %d, want 3", len(assets))
		}
	})

	t.Run("empty ids", func(t *testing.T) {
		client, _ := NewClient("test-key")
		assets, err := client.GetAssetBatch(context.Background(), []string{})

		if err != nil {
			t.Fatalf("GetAssetBatch returned error: %v", err)
		}
		if len(assets) != 0 {
			t.Errorf("len(assets) = %d, want 0", len(assets))
		}
	})
}

func TestAssetTypes(t *testing.T) {
	t.Run("authority type", func(t *testing.T) {
		auth := Authority{
			Address: "authority-address",
			Scopes:  []string{"full"},
		}
		if auth.Address != "authority-address" {
			t.Errorf("Address = %s, unexpected value", auth.Address)
		}
	})

	t.Run("royalty type", func(t *testing.T) {
		royalty := Royalty{
			RoyaltyModel: "creators",
			Percent:      5.5,
			BasisPoints:  550,
			Locked:       true,
		}
		if royalty.BasisPoints != 550 {
			t.Errorf("BasisPoints = %d, want 550", royalty.BasisPoints)
		}
	})

	t.Run("ownership type", func(t *testing.T) {
		ownership := Ownership{
			Owner:          "owner-address",
			OwnershipModel: "single",
			Frozen:         false,
		}
		if ownership.Owner != "owner-address" {
			t.Errorf("Owner = %s, unexpected value", ownership.Owner)
		}
	})

	t.Run("token info with price", func(t *testing.T) {
		info := TokenInfo{
			Symbol:   "BONK",
			Decimals: 5,
			PriceInfo: &Price{
				PricePerToken: 0.00002,
				Currency:      "USDC",
			},
		}
		if info.PriceInfo.PricePerToken != 0.00002 {
			t.Errorf("PricePerToken = %f, unexpected value", info.PriceInfo.PricePerToken)
		}
	})
}

func TestGrouping(t *testing.T) {
	g := Grouping{
		GroupKey:   "collection",
		GroupValue: "collection-mint-address",
	}
	if g.GroupKey != "collection" {
		t.Errorf("GroupKey = %s, want collection", g.GroupKey)
	}
}

func TestSortBy(t *testing.T) {
	sort := SortBy{
		SortBy:        "created",
		SortDirection: "desc",
	}
	if sort.SortBy != "created" {
		t.Errorf("SortBy = %s, want created", sort.SortBy)
	}
}
