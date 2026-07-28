package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk"
)

var (
	_ datasource.DataSource              = &PlatformOAuth2TokenDataSource{}
	_ datasource.DataSourceWithConfigure = &PlatformOAuth2TokenDataSource{}
)

func NewPlatformOAuth2TokenDataSource() datasource.DataSource {
	return &PlatformOAuth2TokenDataSource{}
}

type PlatformOAuth2TokenDataSource struct {
	client *sdk.TerraformProviderNs
}

type PlatformOAuth2TokenDataSourceModel struct {
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	GrantType    types.String `tfsdk:"grant_type"`
	AccessToken  types.String `tfsdk:"access_token"`
	TokenType    types.String `tfsdk:"token_type"`
	ExpiresIn    types.Int64  `tfsdk:"expires_in"`
}

func (d *PlatformOAuth2TokenDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_platform_oauth2_token"
}

func (d *PlatformOAuth2TokenDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Exchanges OAuth2 client credentials for a bearer access token using the RFC 6749 ` + "`client_credentials`" + ` grant.

The token is re-fetched on every Terraform plan and apply. The ` + "`access_token`" + ` is sensitive and stored encrypted in state.

**Note:** OAuth2 clients must be configured in the Netskope tenant UI under
Settings → Security Cloud Platform → OAuth2 Settings before using this data source.`,

		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "OAuth2 client identifier.",
			},
			"client_secret": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "OAuth2 client secret.",
			},
			"grant_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "OAuth2 grant type. Only `client_credentials` is supported. Defaults to `client_credentials`.",
			},
			"access_token": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The issued OAuth2 bearer token.",
			},
			"token_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Token type. Always `Bearer`.",
			},
			"expires_in": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Seconds until the token expires.",
			},
		},
	}
}

func (d *PlatformOAuth2TokenDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sdk.TerraformProviderNs)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *sdk.TerraformProviderNs, got: %T", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *PlatformOAuth2TokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PlatformOAuth2TokenDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	grantType := "client_credentials"
	if !data.GrantType.IsNull() && !data.GrantType.IsUnknown() && data.GrantType.ValueString() != "" {
		grantType = data.GrantType.ValueString()
	}

	tokenResp, err := d.fetchToken(ctx, data.ClientID.ValueString(), data.ClientSecret.ValueString(), grantType)
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch OAuth2 token", err.Error())
		return
	}

	data.GrantType = types.StringValue(grantType)
	data.AccessToken = types.StringValue(tokenResp.AccessToken)
	data.TokenType = types.StringValue(tokenResp.TokenType)
	data.ExpiresIn = types.Int64Value(int64(tokenResp.ExpiresIn))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type oauth2TokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
}

type oauth2TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type oauth2ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (d *PlatformOAuth2TokenDataSource) fetchToken(ctx context.Context, clientID, clientSecret, grantType string) (*oauth2TokenResponse, error) {
	serverURL := d.client.GetServerURL()
	if serverURL == "" {
		return nil, fmt.Errorf("provider server URL is not configured")
	}

	endpoint := serverURL + "/platform/oauth2/token"

	body, err := json.Marshal(oauth2TokenRequest{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		GrantType:    grantType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpClient := d.client.GetHTTPClient()
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		var errResp oauth2ErrorResponse
		if jsonErr := json.Unmarshal(respBody, &errResp); jsonErr == nil && errResp.Error != "" {
			return nil, fmt.Errorf("API error %d: %s — %s", httpResp.StatusCode, errResp.Error, errResp.ErrorDescription)
		}
		return nil, fmt.Errorf("unexpected status %d: %s", httpResp.StatusCode, string(respBody))
	}

	var tokenResp oauth2TokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("API returned empty access_token")
	}

	return &tokenResp, nil
}
