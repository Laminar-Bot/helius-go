package helius

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// Asset represents a digital asset (NFT or token) from the DAS API.
type Asset struct {
	// ID is the asset's unique identifier (mint address).
	ID string `json:"id"`

	// Interface is the asset type (e.g., "V1_NFT", "FungibleToken").
	Interface string `json:"interface"`

	// Content contains metadata and media links.
	Content *AssetContent `json:"content,omitempty"`

	// Authorities lists addresses with authority over the asset.
	Authorities []Authority `json:"authorities,omitempty"`

	// Compression contains compression info for cNFTs.
	Compression *Compression `json:"compression,omitempty"`

	// Grouping contains collection info.
	Grouping []Grouping `json:"grouping,omitempty"`

	// Royalty contains royalty configuration.
	Royalty *Royalty `json:"royalty,omitempty"`

	// Ownership contains current ownership info.
	Ownership *Ownership `json:"ownership,omitempty"`

	// Supply contains supply info for fungible tokens.
	Supply *Supply `json:"supply,omitempty"`

	// TokenInfo contains additional token info.
	TokenInfo *TokenInfo `json:"token_info,omitempty"`

	// Mutable indicates if the asset metadata can be changed.
	Mutable bool `json:"mutable"`

	// Burnt indicates if the asset has been burned.
	Burnt bool `json:"burnt"`
}

// AssetContent contains asset metadata and media.
type AssetContent struct {
	Schema   string                 `json:"$schema,omitempty"`
	JSONUri  string                 `json:"json_uri,omitempty"`
	Files    []AssetFile            `json:"files,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Links    map[string]string      `json:"links,omitempty"`
}

// AssetFile represents a file associated with an asset.
type AssetFile struct {
	URI  string `json:"uri"`
	Mime string `json:"mime,omitempty"`
	CDN  bool   `json:"cdn,omitempty"`
}

// Authority represents an authority over an asset.
type Authority struct {
	Address string   `json:"address"`
	Scopes  []string `json:"scopes"`
}

// Compression contains compression info for compressed NFTs.
type Compression struct {
	Eligible    bool   `json:"eligible"`
	Compressed  bool   `json:"compressed"`
	DataHash    string `json:"data_hash,omitempty"`
	CreatorHash string `json:"creator_hash,omitempty"`
	AssetHash   string `json:"asset_hash,omitempty"`
	Tree        string `json:"tree,omitempty"`
	Seq         int64  `json:"seq,omitempty"`
	LeafID      int64  `json:"leaf_id,omitempty"`
}

// Grouping represents collection grouping.
type Grouping struct {
	GroupKey   string `json:"group_key"`
	GroupValue string `json:"group_value"`
}

// Royalty contains royalty configuration.
type Royalty struct {
	RoyaltyModel        string  `json:"royalty_model"`
	Target              string  `json:"target,omitempty"`
	Percent             float64 `json:"percent"`
	BasisPoints         int     `json:"basis_points"`
	PrimarySaleHappened bool    `json:"primary_sale_happened"`
	Locked              bool    `json:"locked"`
}

// Ownership contains current ownership information.
type Ownership struct {
	Frozen         bool   `json:"frozen"`
	Delegated      bool   `json:"delegated"`
	Delegate       string `json:"delegate,omitempty"`
	OwnershipModel string `json:"ownership_model"`
	Owner          string `json:"owner"`
}

// Supply contains token supply information.
type Supply struct {
	PrintMaxSupply  int64 `json:"print_max_supply"`
	PrintCurrentSup int64 `json:"print_current_supply"`
	EditionNonce    int   `json:"edition_nonce,omitempty"`
}

// TokenInfo contains additional token information.
type TokenInfo struct {
	Symbol                 string `json:"symbol,omitempty"`
	Balance                int64  `json:"balance,omitempty"`
	Supply                 int64  `json:"supply,omitempty"`
	Decimals               int    `json:"decimals,omitempty"`
	TokenProgram           string `json:"token_program,omitempty"`
	AssociatedTokenAddress string `json:"associated_token_address,omitempty"`
	PriceInfo              *Price `json:"price_info,omitempty"`
}

// Price contains price information.
type Price struct {
	PricePerToken float64 `json:"price_per_token"`
	TotalPrice    float64 `json:"total_price,omitempty"`
	Currency      string  `json:"currency,omitempty"`
}

// AssetsPage represents a paginated response of assets.
type AssetsPage struct {
	Total         int     `json:"total"`
	Limit         int     `json:"limit"`
	Page          int     `json:"page,omitempty"`
	Cursor        string  `json:"cursor,omitempty"`
	Items         []Asset `json:"items"`
	NativeBalance *Balance `json:"nativeBalance,omitempty"`
}

// Balance represents a native SOL balance.
type Balance struct {
	Lamports            int64   `json:"lamports"`
	PricePerSOL         float64 `json:"price_per_sol,omitempty"`
	TotalPrice          float64 `json:"total_price,omitempty"`
}

// GetAssetOptions configures the GetAsset request.
type GetAssetOptions struct {
	ShowFungible           bool `json:"showFungible,omitempty"`
	ShowUnverifiedCollect  bool `json:"showUnverifiedCollections,omitempty"`
	ShowCollectionMetadata bool `json:"showCollectionMetadata,omitempty"`
	ShowGrandTotal         bool `json:"showGrandTotal,omitempty"`
	ShowInscription        bool `json:"showInscription,omitempty"`
}

// GetAsset fetches a single asset by its ID (mint address).
func (c *Client) GetAsset(ctx context.Context, id string) (*Asset, error) {
	if id == "" {
		return nil, &APIError{
			StatusCode: 400,
			Message:    "asset ID is required",
			Path:       "/assets",
		}
	}

	reqBody := map[string]interface{}{
		"id": id,
	}

	body, err := c.doPost(ctx, "/assets", reqBody)
	if err != nil {
		return nil, err
	}

	var asset Asset
	if err := json.Unmarshal(body, &asset); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	c.logger.Debug("fetched asset", "id", id, "interface", asset.Interface)

	return &asset, nil
}

// AssetsByOwnerOptions configures the GetAssetsByOwner request.
type AssetsByOwnerOptions struct {
	Page                   int  `json:"page,omitempty"`
	Limit                  int  `json:"limit,omitempty"`
	Cursor                 string `json:"cursor,omitempty"`
	Before                 string `json:"before,omitempty"`
	After                  string `json:"after,omitempty"`
	ShowFungible           bool `json:"showFungible,omitempty"`
	ShowNativeBalance      bool `json:"showNativeBalance,omitempty"`
	ShowUnverifiedCollect  bool `json:"showUnverifiedCollections,omitempty"`
	ShowCollectionMetadata bool `json:"showCollectionMetadata,omitempty"`
	ShowGrandTotal         bool `json:"showGrandTotal,omitempty"`
	ShowZeroBalance        bool `json:"showZeroBalance,omitempty"`
	SortBy                 *SortBy `json:"sortBy,omitempty"`
}

// SortBy configures sorting for asset queries.
type SortBy struct {
	SortBy        string `json:"sortBy"`        // "created", "updated", "recent_action"
	SortDirection string `json:"sortDirection"` // "asc", "desc"
}

// GetAssetsByOwner fetches all assets owned by an address.
func (c *Client) GetAssetsByOwner(ctx context.Context, ownerAddress string, opts *AssetsByOwnerOptions) (*AssetsPage, error) {
	if ownerAddress == "" {
		return nil, &APIError{
			StatusCode: 400,
			Message:    "owner address is required",
			Path:       "/assets",
		}
	}

	reqBody := map[string]interface{}{
		"ownerAddress": ownerAddress,
	}

	if opts != nil {
		if opts.Page > 0 {
			reqBody["page"] = opts.Page
		}
		if opts.Limit > 0 {
			reqBody["limit"] = opts.Limit
		}
		if opts.Cursor != "" {
			reqBody["cursor"] = opts.Cursor
		}
		if opts.Before != "" {
			reqBody["before"] = opts.Before
		}
		if opts.After != "" {
			reqBody["after"] = opts.After
		}

		displayOpts := map[string]bool{}
		if opts.ShowFungible {
			displayOpts["showFungible"] = true
		}
		if opts.ShowNativeBalance {
			displayOpts["showNativeBalance"] = true
		}
		if opts.ShowUnverifiedCollect {
			displayOpts["showUnverifiedCollections"] = true
		}
		if opts.ShowCollectionMetadata {
			displayOpts["showCollectionMetadata"] = true
		}
		if opts.ShowGrandTotal {
			displayOpts["showGrandTotal"] = true
		}
		if opts.ShowZeroBalance {
			displayOpts["showZeroBalance"] = true
		}
		if len(displayOpts) > 0 {
			reqBody["displayOptions"] = displayOpts
		}

		if opts.SortBy != nil {
			reqBody["sortBy"] = opts.SortBy
		}
	}

	body, err := c.doPost(ctx, "/assets", reqBody)
	if err != nil {
		return nil, err
	}

	var page AssetsPage
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	c.logger.Debug("fetched assets by owner",
		"owner", ownerAddress,
		"total", page.Total,
		"returned", len(page.Items),
	)

	return &page, nil
}

// SearchAssetsOptions configures the SearchAssets request.
type SearchAssetsOptions struct {
	Page                   int      `json:"page,omitempty"`
	Limit                  int      `json:"limit,omitempty"`
	Cursor                 string   `json:"cursor,omitempty"`
	OwnerAddress           string   `json:"ownerAddress,omitempty"`
	CreatorAddress         string   `json:"creatorAddress,omitempty"`
	CreatorVerified        *bool    `json:"creatorVerified,omitempty"`
	AuthorityAddress       string   `json:"authorityAddress,omitempty"`
	GroupKey               string   `json:"groupKey,omitempty"`
	GroupValue             string   `json:"groupValue,omitempty"`
	Delegate               string   `json:"delegate,omitempty"`
	Frozen                 *bool    `json:"frozen,omitempty"`
	Supply                 *int64   `json:"supply,omitempty"`
	SupplyMint             string   `json:"supplyMint,omitempty"`
	Compressed             *bool    `json:"compressed,omitempty"`
	Compressible           *bool    `json:"compressible,omitempty"`
	RoyaltyTargetType      string   `json:"royaltyTargetType,omitempty"`
	RoyaltyTarget          string   `json:"royaltyTarget,omitempty"`
	RoyaltyAmount          *int     `json:"royaltyAmount,omitempty"`
	Burnt                  *bool    `json:"burnt,omitempty"`
	Interface              string   `json:"interface,omitempty"`
	TokenType              string   `json:"tokenType,omitempty"`
	OwnerType              string   `json:"ownerType,omitempty"`
	SpecificationVersion   string   `json:"specificationVersion,omitempty"`
	ShowFungible           bool     `json:"showFungible,omitempty"`
	ShowCollectionMetadata bool     `json:"showCollectionMetadata,omitempty"`
	SortBy                 *SortBy  `json:"sortBy,omitempty"`
	JsonUri                string   `json:"jsonUri,omitempty"`
}

// SearchAssets searches for assets matching the given criteria.
func (c *Client) SearchAssets(ctx context.Context, opts *SearchAssetsOptions) (*AssetsPage, error) {
	if opts == nil {
		return nil, &APIError{
			StatusCode: 400,
			Message:    "search options are required",
			Path:       "/assets/search",
		}
	}

	reqBody := make(map[string]interface{})

	if opts.Page > 0 {
		reqBody["page"] = opts.Page
	}
	if opts.Limit > 0 {
		reqBody["limit"] = opts.Limit
	}
	if opts.Cursor != "" {
		reqBody["cursor"] = opts.Cursor
	}
	if opts.OwnerAddress != "" {
		reqBody["ownerAddress"] = opts.OwnerAddress
	}
	if opts.CreatorAddress != "" {
		reqBody["creatorAddress"] = opts.CreatorAddress
	}
	if opts.CreatorVerified != nil {
		reqBody["creatorVerified"] = *opts.CreatorVerified
	}
	if opts.AuthorityAddress != "" {
		reqBody["authorityAddress"] = opts.AuthorityAddress
	}
	if opts.GroupKey != "" {
		reqBody["groupKey"] = opts.GroupKey
	}
	if opts.GroupValue != "" {
		reqBody["groupValue"] = opts.GroupValue
	}
	if opts.Delegate != "" {
		reqBody["delegate"] = opts.Delegate
	}
	if opts.Frozen != nil {
		reqBody["frozen"] = *opts.Frozen
	}
	if opts.Compressed != nil {
		reqBody["compressed"] = *opts.Compressed
	}
	if opts.Burnt != nil {
		reqBody["burnt"] = *opts.Burnt
	}
	if opts.Interface != "" {
		reqBody["interface"] = opts.Interface
	}
	if opts.TokenType != "" {
		reqBody["tokenType"] = opts.TokenType
	}
	if opts.JsonUri != "" {
		reqBody["jsonUri"] = opts.JsonUri
	}
	if opts.SortBy != nil {
		reqBody["sortBy"] = opts.SortBy
	}

	body, err := c.doPost(ctx, "/assets/search", reqBody)
	if err != nil {
		return nil, err
	}

	var page AssetsPage
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	c.logger.Debug("searched assets", "total", page.Total, "returned", len(page.Items))

	return &page, nil
}

// GetAssetBatch fetches multiple assets by their IDs.
func (c *Client) GetAssetBatch(ctx context.Context, ids []string) ([]Asset, error) {
	if len(ids) == 0 {
		return []Asset{}, nil
	}

	reqBody := map[string]interface{}{
		"ids": ids,
	}

	body, err := c.doPost(ctx, "/assets/batch", reqBody)
	if err != nil {
		return nil, err
	}

	var assets []Asset
	if err := json.Unmarshal(body, &assets); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	c.logger.Debug("fetched asset batch", "requested", len(ids), "returned", len(assets))

	return assets, nil
}

// doPost with bytes.Buffer for proper body handling
func (c *Client) doPostJSON(ctx context.Context, path string, reqBody interface{}) ([]byte, error) {
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	return c.doRequest(ctx, "POST", path, bytes.NewReader(jsonBody))
}
