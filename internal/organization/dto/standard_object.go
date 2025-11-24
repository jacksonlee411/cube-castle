package dto

// StandardObjectFilter 过滤条件
type StandardObjectFilter struct {
	Code     *string `json:"code"`
	Status   *string `json:"status"`
	AsOfDate *string `json:"asOfDate"`
}

// StandardObjectConnection GraphQL 分页封装
type StandardObjectConnection struct {
	DataField       []StandardObject `json:"data"`
	PaginationField PaginationInfo   `json:"pagination"`
}

func (c StandardObjectConnection) Data() []StandardObject { return c.DataField }
func (c StandardObjectConnection) Pagination() PaginationInfo {
	return c.PaginationField
}

// StandardObject 聚合
type StandardObject struct {
	KernelField  StandardObjectKernel  `json:"kernel"`
	VersionField StandardObjectVersion `json:"version"`
	LinksField   []StandardObjectLink  `json:"links"`
}

func (s StandardObject) Kernel() StandardObjectKernel   { return s.KernelField }
func (s StandardObject) Version() StandardObjectVersion { return s.VersionField }
func (s StandardObject) Links() []StandardObjectLink    { return s.LinksField }

// StandardObjectKernel SOM Kernel
type StandardObjectKernel struct {
	ObjectTypeField         string            `json:"objectType"`
	CodeField               string            `json:"code"`
	DisplayNameField        string            `json:"displayName"`
	TenantCodeField         string            `json:"tenantCode"`
	StatusField             string            `json:"status"`
	LabelsField             map[string]string `json:"labels"`
	SchemaVersionField      *string           `json:"schemaVersion"`
	DataClassificationField *string           `json:"dataClassification"`
	RetentionPolicyField    *string           `json:"retentionPolicy"`
	CreatedByField          *string           `json:"createdBy"`
	CreatedAtField          string            `json:"createdAt"`
	UpdatedAtField          string            `json:"updatedAt"`
}

// StandardObjectVersion SOM 版本
type StandardObjectVersion struct {
	VersionCodeField   string         `json:"versionCode"`
	EffectiveDateField string         `json:"effectiveDate"`
	EndDateField       *string        `json:"endDate"`
	IsCurrentField     bool           `json:"isCurrent"`
	PayloadField       map[string]any `json:"payload"`
	AuditTrailField    map[string]any `json:"auditTrail"`
	ChecksumField      *string        `json:"checksum"`
	CreatedAtField     *string        `json:"createdAt"`
	UpdatedAtField     *string        `json:"updatedAt"`
}

// StandardObjectLink SOM Link
type StandardObjectLink struct {
	LinkTypeField   string         `json:"linkType"`
	SourceCodeField string         `json:"sourceCode"`
	TargetCodeField string         `json:"targetCode"`
	AttributesField map[string]any `json:"attributes"`
}
