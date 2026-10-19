package kustolang

// Code generated from the pinned upstream Kusto.Language sources; DO NOT EDIT.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type engineCommandsSet struct {
	ShowDatabase                                          *CommandSymbol
	ShowDatabaseDetails                                   *CommandSymbol
	ShowDatabaseIdentity                                  *CommandSymbol
	ShowDatabasePolicies                                  *CommandSymbol
	ShowDatabaseDataStats                                 *CommandSymbol
	ShowDatabaseMetadata                                  *CommandSymbol
	ShowClusterDatabases                                  *CommandSymbol
	ShowClusterDatabasesDetails                           *CommandSymbol
	ShowClusterDatabasesIdentity                          *CommandSymbol
	ShowClusterDatabasesPolicies                          *CommandSymbol
	ShowClusterDatabasesDataStats                         *CommandSymbol
	ShowClusterDatabasesMetadata                          *CommandSymbol
	CreateDatabase                                        *CommandSymbol
	AttachDatabase                                        *CommandSymbol
	DetachDatabase                                        *CommandSymbol
	AlterDatabasePrettyName                               *CommandSymbol
	DropDatabasePrettyName                                *CommandSymbol
	AlterDatabasePersistMetadata                          *CommandSymbol
	SetAccess                                             *CommandSymbol
	ShowDatabaseSchema                                    *CommandSymbol
	ShowDatabaseSchemaAsJson                              *CommandSymbol
	ShowDatabaseSchemaAsCslScript                         *CommandSymbol
	ShowDatabaseCslSchema                                 *CommandSymbol
	ShowDatabaseSchemaViolations                          *CommandSymbol
	ShowDatabasesSchema                                   *CommandSymbol
	ShowDatabasesSchemaAsJson                             *CommandSymbol
	CreateDatabaseIngestionMapping                        *CommandSymbol
	CreateOrAlterDatabaseIngestionMapping                 *CommandSymbol
	AlterDatabaseIngestionMapping                         *CommandSymbol
	AlterMergeDatabaseIngestionMapping                    *CommandSymbol
	ShowDatabaseIngestionMappings                         *CommandSymbol
	ShowIngestionMappings                                 *CommandSymbol
	DropDatabaseIngestionMapping                          *CommandSymbol
	ShowTables                                            *CommandSymbol
	ShowTable                                             *CommandSymbol
	ShowTablesDetails                                     *CommandSymbol
	ShowTableDetails                                      *CommandSymbol
	ShowTableCslSchema                                    *CommandSymbol
	ShowTableSchemaAsJson                                 *CommandSymbol
	CreateTable                                           *CommandSymbol
	CreateTableBasedOnAnother                             *CommandSymbol
	CreateMergeTable                                      *CommandSymbol
	DefineTable                                           *CommandSymbol
	CreateTables                                          *CommandSymbol
	CreateMergeTables                                     *CommandSymbol
	DefineTables                                          *CommandSymbol
	AlterTable                                            *CommandSymbol
	AlterMergeTable                                       *CommandSymbol
	AlterTableDocString                                   *CommandSymbol
	AlterTableFolder                                      *CommandSymbol
	RenameTable                                           *CommandSymbol
	RenameTables                                          *CommandSymbol
	UndoDropExtentContainer                               *CommandSymbol
	DropTable                                             *CommandSymbol
	UndoDropTable                                         *CommandSymbol
	DropTables                                            *CommandSymbol
	CreateTableIngestionMapping                           *CommandSymbol
	CreateOrAlterTableIngestionMapping                    *CommandSymbol
	AlterTableIngestionMapping                            *CommandSymbol
	AlterMergeTableIngestionMapping                       *CommandSymbol
	ShowTableIngestionMappings                            *CommandSymbol
	ShowTableIngestionMapping                             *CommandSymbol
	DropTableIngestionMapping                             *CommandSymbol
	RenameColumn                                          *CommandSymbol
	RenameColumns                                         *CommandSymbol
	AlterColumnType                                       *CommandSymbol
	DropColumn                                            *CommandSymbol
	DropTableColumns                                      *CommandSymbol
	AlterTableColumnDocStrings                            *CommandSymbol
	AlterMergeTableColumnDocStrings                       *CommandSymbol
	ShowFunctions                                         *CommandSymbol
	ShowFunction                                          *CommandSymbol
	CreateFunction                                        *CommandSymbol
	AlterFunction                                         *CommandSymbol
	CreateOrAlterFunction                                 *CommandSymbol
	DropFunction                                          *CommandSymbol
	DropFunctions                                         *CommandSymbol
	AlterFunctionDocString                                *CommandSymbol
	AlterFunctionFolder                                   *CommandSymbol
	ShowExternalTables                                    *CommandSymbol
	ShowExternalTable                                     *CommandSymbol
	ShowExternalTablesDetails                             *CommandSymbol
	ShowExternalTableDetails                              *CommandSymbol
	ShowExternalTableCslSchema                            *CommandSymbol
	ShowExternalTableSchema                               *CommandSymbol
	ShowExternalTableArtifacts                            *CommandSymbol
	DropExternalTable                                     *CommandSymbol
	CreateStorageExternalTable                            *CommandSymbol
	AlterStorageExternalTable                             *CommandSymbol
	CreateOrAlterStorageExternalTable                     *CommandSymbol
	CreateSqlExternalTable                                *CommandSymbol
	AlterSqlExternalTable                                 *CommandSymbol
	CreateOrAlterSqlExternalTable                         *CommandSymbol
	CreateExternalTableMapping                            *CommandSymbol
	SetExternalTableAdmins                                *CommandSymbol
	AddExternalTableAdmins                                *CommandSymbol
	DropExternalTableAdmins                               *CommandSymbol
	AlterExternalTableDocString                           *CommandSymbol
	AlterExternalTableFolder                              *CommandSymbol
	ShowExternalTablePrincipals                           *CommandSymbol
	ShowFabric                                            *CommandSymbol
	AlterExternalTableMapping                             *CommandSymbol
	ShowExternalTableMappings                             *CommandSymbol
	ShowExternalTableMapping                              *CommandSymbol
	DropExternalTableMapping                              *CommandSymbol
	ShowWorkloadGroups                                    *CommandSymbol
	ShowWorkloadGroup                                     *CommandSymbol
	CreateOrAleterWorkloadGroup                           *CommandSymbol
	AlterMergeWorkloadGroup                               *CommandSymbol
	DropWorkloadGroup                                     *CommandSymbol
	ShowDatabasePolicyCaching                             *CommandSymbol
	ShowTablePolicyCaching                                *CommandSymbol
	ShowTableStarPolicyCaching                            *CommandSymbol
	ShowColumnPolicyCaching                               *CommandSymbol
	ShowMaterializedViewPolicyCaching                     *CommandSymbol
	ShowGraphModelPolicyCaching                           *CommandSymbol
	ShowGraphModelStarPolicyCaching                       *CommandSymbol
	ShowClusterPolicyCaching                              *CommandSymbol
	AlterDatabasePolicyCaching                            *CommandSymbol
	AlterTablePolicyCaching                               *CommandSymbol
	AlterTablesPolicyCaching                              *CommandSymbol
	AlterColumnPolicyCaching                              *CommandSymbol
	AlterMaterializedViewPolicyCaching                    *CommandSymbol
	AlterGraphModelPolicyCaching                          *CommandSymbol
	AlterClusterPolicyCaching                             *CommandSymbol
	DeleteDatabasePolicyCaching                           *CommandSymbol
	DeleteTablePolicyCaching                              *CommandSymbol
	DeleteColumnPolicyCaching                             *CommandSymbol
	DeleteMaterializedViewPolicyCaching                   *CommandSymbol
	DeleteGraphModelPolicyCaching                         *CommandSymbol
	DeleteClusterPolicyCaching                            *CommandSymbol
	ShowTablePolicyChangeDataCapture                      *CommandSymbol
	AlterTablePolicyChangeDataCapture                     *CommandSymbol
	AlterTablesPolicyChangeDataCapture                    *CommandSymbol
	DeleteTablePolicyChangeDataCapture                    *CommandSymbol
	ShowTablePolicyIngestionTime                          *CommandSymbol
	ShowTableStarPolicyIngestionTime                      *CommandSymbol
	AlterTablePolicyIngestionTime                         *CommandSymbol
	AlterTablesPolicyIngestionTime                        *CommandSymbol
	DeleteTablePolicyIngestionTime                        *CommandSymbol
	ShowDatabasePolicyRowId                               *CommandSymbol
	AlterDatabasePolicyRowId                              *CommandSymbol
	DeleteDatabasePolicyRowId                             *CommandSymbol
	ShowTablePolicyRowId                                  *CommandSymbol
	AlterTablePolicyRowId                                 *CommandSymbol
	AlterTablesPolicyRowId                                *CommandSymbol
	DeleteTablePolicyRowId                                *CommandSymbol
	ShowTablePolicyRetention                              *CommandSymbol
	ShowTableStarPolicyRetention                          *CommandSymbol
	ShowGraphPolicyRetention                              *CommandSymbol
	ShowGraphStarPolicyRetention                          *CommandSymbol
	ShowDatabasePolicyRetention                           *CommandSymbol
	AlterTablePolicyRetention                             *CommandSymbol
	AlterMaterializedViewPolicyRetention                  *CommandSymbol
	AlterDatabasePolicyRetention                          *CommandSymbol
	AlterGraphModelPolicyRetention                        *CommandSymbol
	AlterTablesPolicyRetention                            *CommandSymbol
	AlterMergeTablePolicyRetention                        *CommandSymbol
	AlterMergeMaterializedViewPolicyRetention             *CommandSymbol
	AlterMergeDatabasePolicyRetention                     *CommandSymbol
	DeleteTablePolicyRetention                            *CommandSymbol
	DeleteDatabasePolicyRetention                         *CommandSymbol
	ShowDatabasePolicyHardRetentionViolations             *CommandSymbol
	ShowDatabasePolicySoftRetentionViolations             *CommandSymbol
	ShowTablePolicyRowLevelSecurity                       *CommandSymbol
	ShowTableStarPolicyRowLevelSecurity                   *CommandSymbol
	AlterTablePolicyRowLevelSecurity                      *CommandSymbol
	DeleteTablePolicyRowLevelSecurity                     *CommandSymbol
	ShowMaterializedViewPolicyRowLevelSecurity            *CommandSymbol
	AlterMaterializedViewPolicyRowLevelSecurity           *CommandSymbol
	DeleteMaterializedViewPolicyRowLevelSecurity          *CommandSymbol
	ShowExternalTablePolicyRowLevelSecurity               *CommandSymbol
	ShowExternalTableStarPolicyRowLevelSecurity           *CommandSymbol
	AlterExternalTablePolicyRowLevelSecurity              *CommandSymbol
	DeleteExternalTablePolicyRowLevelSecurity             *CommandSymbol
	ShowTablePolicyRowOrder                               *CommandSymbol
	ShowTableStarPolicyRowOrder                           *CommandSymbol
	AlterTablePolicyRowOrder                              *CommandSymbol
	AlterTablesPolicyRowOrder                             *CommandSymbol
	AlterMergeTablePolicyRowOrder                         *CommandSymbol
	DeleteTablePolicyRowOrder                             *CommandSymbol
	ShowTablePolicyUpdate                                 *CommandSymbol
	ShowTableStarPolicyUpdate                             *CommandSymbol
	AlterTablePolicyUpdate                                *CommandSymbol
	AlterMergeTablePolicyUpdate                           *CommandSymbol
	DeleteTablePolicyUpdate                               *CommandSymbol
	ShowClusterPolicyIngestionBatching                    *CommandSymbol
	ShowDatabasePolicyIngestionBatching                   *CommandSymbol
	ShowTablePolicyIngestionBatching                      *CommandSymbol
	ShowTableStarPolicyIngestionBatching                  *CommandSymbol
	AlterClusterPolicyIngestionBatching                   *CommandSymbol
	AlterMergeClusterPolicyIngestionBatching              *CommandSymbol
	AlterDatabasePolicyIngestionBatching                  *CommandSymbol
	AlterMergeDatabasePolicyIngestionBatching             *CommandSymbol
	AlterTablePolicyIngestionBatching                     *CommandSymbol
	AlterMergeTablePolicyIngestionBatching                *CommandSymbol
	AlterTablesPolicyIngestionBatching                    *CommandSymbol
	DeleteDatabasePolicyIngestionBatching                 *CommandSymbol
	DeleteTablePolicyIngestionBatching                    *CommandSymbol
	ShowDatabasePolicyEncoding                            *CommandSymbol
	ShowTablePolicyEncoding                               *CommandSymbol
	ShowColumnPolicyEncoding                              *CommandSymbol
	AlterDatabasePolicyEncoding                           *CommandSymbol
	AlterTablePolicyEncoding                              *CommandSymbol
	AlterTableColumnsPolicyEncoding                       *CommandSymbol
	AlterColumnPolicyEncoding                             *CommandSymbol
	AlterColumnsPolicyEncodingByQuery                     *CommandSymbol
	AlterColumnPolicyEncodingType                         *CommandSymbol
	AlterMergeDatabasePolicyEncoding                      *CommandSymbol
	AlterMergeTablePolicyEncoding                         *CommandSymbol
	AlterMergeColumnPolicyEncoding                        *CommandSymbol
	DeleteDatabasePolicyEncoding                          *CommandSymbol
	DeleteTablePolicyEncoding                             *CommandSymbol
	DeleteColumnPolicyEncoding                            *CommandSymbol
	ShowDatabasePolicyMerge                               *CommandSymbol
	ShowTablePolicyMerge                                  *CommandSymbol
	ShowTableStarPolicyMerge                              *CommandSymbol
	AlterDatabasePolicyMerge                              *CommandSymbol
	AlterTablePolicyMerge                                 *CommandSymbol
	AlterTablesPolicyMerge                                *CommandSymbol
	AlterMergeDatabasePolicyMerge                         *CommandSymbol
	AlterMergeTablePolicyMerge                            *CommandSymbol
	AlterMergeMaterializedViewPolicyMerge                 *CommandSymbol
	DeleteDatabasePolicyMerge                             *CommandSymbol
	DeleteTablePolicyMerge                                *CommandSymbol
	ShowExternalTablePolicyQueryAcceleration              *CommandSymbol
	ShowExternalTablesPolicyQueryAcceleration             *CommandSymbol
	AlterExternalTablePolicyQueryAcceleration             *CommandSymbol
	AlterMergeExternalTablePolicyQueryAcceleration        *CommandSymbol
	DeleteExternalTablePolicyQueryAcceleration            *CommandSymbol
	ShowExternalTableQueryAccelerationStatatistics        *CommandSymbol
	ShowExternalTablesQueryAccelerationStatatistics       *CommandSymbol
	AlterTablePolicyMirroring                             *CommandSymbol
	AlterMergeTablePolicyMirroring                        *CommandSymbol
	AlterTablePolicyMirroringWithJson                     *CommandSymbol
	AlterMergeTablePolicyMirroringWithJson                *CommandSymbol
	DeleteTablePolicyMirroring                            *CommandSymbol
	ShowTablePolicyMirroring                              *CommandSymbol
	ShowTableStarPolicyMirroring                          *CommandSymbol
	CreateMirroringTemplate                               *CommandSymbol
	CreateOrAlterMirroringTemplate                        *CommandSymbol
	AlterMirroringTemplate                                *CommandSymbol
	AlterMergeMirroringTemplate                           *CommandSymbol
	DeleteMirroringTemplate                               *CommandSymbol
	ShowMirroringTemplate                                 *CommandSymbol
	ShowMirroringTemplates                                *CommandSymbol
	ApplyMirroringTemplate                                *CommandSymbol
	ShowTablePolicyPartitioning                           *CommandSymbol
	ShowTableStarPolicyPartitioning                       *CommandSymbol
	AlterTablePolicyPartitioning                          *CommandSymbol
	AlterMergeTablePolicyPartitioning                     *CommandSymbol
	AlterMaterializedViewPolicyPartitioning               *CommandSymbol
	AlterMergeMaterializedViewPolicyPartitioning          *CommandSymbol
	DeleteTablePolicyPartitioning                         *CommandSymbol
	DeleteMaterializedViewPolicyPartitioning              *CommandSymbol
	ShowTablePolicyRestrictedViewAccess                   *CommandSymbol
	ShowTableStarPolicyRestrictedViewAccess               *CommandSymbol
	AlterTablePolicyRestrictedViewAccess                  *CommandSymbol
	AlterTablesPolicyRestrictedViewAccess                 *CommandSymbol
	DeleteTablePolicyRestrictedViewAccess                 *CommandSymbol
	ShowClusterPolicyRowStore                             *CommandSymbol
	AlterClusterPolicyRowStore                            *CommandSymbol
	AlterMergeClusterPolicyRowStore                       *CommandSymbol
	DeleteClusterPolicyRowStore                           *CommandSymbol
	ShowClusterPolicySandbox                              *CommandSymbol
	AlterClusterPolicySandbox                             *CommandSymbol
	DeleteClusterPolicySandbox                            *CommandSymbol
	ShowClusterSandboxesStats                             *CommandSymbol
	ShowDatabasePolicySharding                            *CommandSymbol
	ShowTablePolicySharding                               *CommandSymbol
	ShowTableStarPolicySharding                           *CommandSymbol
	AlterDatabasePolicySharding                           *CommandSymbol
	AlterTablePolicySharding                              *CommandSymbol
	AlterMergeDatabasePolicySharding                      *CommandSymbol
	AlterMergeTablePolicySharding                         *CommandSymbol
	DeleteDatabasePolicySharding                          *CommandSymbol
	DeleteTablePolicySharding                             *CommandSymbol
	AlterClusterPolicySharding                            *CommandSymbol
	AlterMergeClusterPolicySharding                       *CommandSymbol
	DeleteClusterPolicySharding                           *CommandSymbol
	ShowClusterPolicySharding                             *CommandSymbol
	ShowDatabasePolicyShardsGrouping                      *CommandSymbol
	AlterDatabasePolicyShardsGrouping                     *CommandSymbol
	AlterMergeDatabasePolicyShardsGrouping                *CommandSymbol
	DeleteDatabasePolicyShardsGrouping                    *CommandSymbol
	ShowDatabasePolicyStreamingIngestion                  *CommandSymbol
	ShowTablePolicyStreamingIngestion                     *CommandSymbol
	ShowClusterPolicyStreamingIngestion                   *CommandSymbol
	AlterDatabasePolicyStreamingIngestion                 *CommandSymbol
	AlterMergeDatabasePolicyStreamingIngestion            *CommandSymbol
	AlterTablePolicyStreamingIngestion                    *CommandSymbol
	AlterMergeTablePolicyStreamingIngestion               *CommandSymbol
	AlterClusterPolicyStreamingIngestion                  *CommandSymbol
	AlterMergeClusterPolicyStreamingIngestion             *CommandSymbol
	DeleteDatabasePolicyStreamingIngestion                *CommandSymbol
	DeleteTablePolicyStreamingIngestion                   *CommandSymbol
	DeleteClusterPolicyStreamingIngestion                 *CommandSymbol
	ShowDatabasePolicyManagedIdentity                     *CommandSymbol
	ShowClusterPolicyManagedIdentity                      *CommandSymbol
	AlterDatabasePolicyManagedIdentity                    *CommandSymbol
	AlterMergeDatabasePolicyManagedIdentity               *CommandSymbol
	AlterClusterPolicyManagedIdentity                     *CommandSymbol
	AlterMergeClusterPolicyManagedIdentity                *CommandSymbol
	DeleteDatabasePolicyManagedIdentity                   *CommandSymbol
	DeleteClusterPolicyManagedIdentity                    *CommandSymbol
	ShowTablePolicyAutoDelete                             *CommandSymbol
	AlterTablePolicyAutoDelete                            *CommandSymbol
	DeleteTablePolicyAutoDelete                           *CommandSymbol
	ShowClusterPolicyCallout                              *CommandSymbol
	AlterClusterPolicyCallout                             *CommandSymbol
	AlterMergeClusterPolicyCallout                        *CommandSymbol
	DeleteClusterPolicyCallout                            *CommandSymbol
	ShowClusterPolicyCapacity                             *CommandSymbol
	AlterClusterPolicyCapacity                            *CommandSymbol
	AlterMergeClusterPolicyCapacity                       *CommandSymbol
	ShowClusterPolicyMaterializedViewPrioritization       *CommandSymbol
	AlterClusterPolicyMaterializedViewPrioritization      *CommandSymbol
	AlterMergeClusterPolicyMaterializedViewPrioritization *CommandSymbol
	DropClusterPolicyMaterializedViewPrioritization       *CommandSymbol
	ShowClusterPolicyRequestClassification                *CommandSymbol
	AlterClusterPolicyRequestClassification               *CommandSymbol
	AlterMergeClusterPolicyRequestClassification          *CommandSymbol
	DeleteClusterPolicyRequestClassification              *CommandSymbol
	ShowClusterPolicyMultiDatabaseAdmins                  *CommandSymbol
	AlterClusterPolicyMultiDatabaseAdmins                 *CommandSymbol
	AlterMergeClusterPolicyMultiDatabaseAdmins            *CommandSymbol
	ShowDatabasePolicyDiagnostics                         *CommandSymbol
	ShowClusterPolicyDiagnostics                          *CommandSymbol
	AlterDatabasePolicyDiagnostics                        *CommandSymbol
	AlterMergeDatabasePolicyDiagnostics                   *CommandSymbol
	AlterClusterPolicyDiagnostics                         *CommandSymbol
	AlterMergeClusterPolicyDiagnostics                    *CommandSymbol
	DeleteDatabasePolicyDiagnostics                       *CommandSymbol
	ShowClusterPolicyQueryWeakConsistency                 *CommandSymbol
	AlterClusterPolicyQueryWeakConsistency                *CommandSymbol
	AlterMergeClusterPolicyQueryWeakConsistency           *CommandSymbol
	ShowTablePolicyExtentTagsRetention                    *CommandSymbol
	ShowTableStarPolicyExtentTagsRetention                *CommandSymbol
	ShowDatabasePolicyExtentTagsRetention                 *CommandSymbol
	AlterTablePolicyExtentTagsRetention                   *CommandSymbol
	AlterDatabasePolicyExtentTagsRetention                *CommandSymbol
	DeleteTablePolicyExtentTagsRetention                  *CommandSymbol
	DeleteDatabasePolicyExtentTagsRetention               *CommandSymbol
	ShowPrincipalRoles                                    *CommandSymbol
	ShowDatabasePrincipalRoles                            *CommandSymbol
	ShowTablePrincipalRoles                               *CommandSymbol
	ShowGraphModelPrincipalRoles                          *CommandSymbol
	ShowExternalTablesPrincipalRoles                      *CommandSymbol
	ShowFunctionPrincipalRoles                            *CommandSymbol
	ShowClusterPrincipalRoles                             *CommandSymbol
	ShowClusterPrincipals                                 *CommandSymbol
	ShowDatabasePrincipals                                *CommandSymbol
	ShowTablePrincipals                                   *CommandSymbol
	ShowGraphModelPrincipals                              *CommandSymbol
	ShowFunctionPrincipals                                *CommandSymbol
	AddClusterRole                                        *CommandSymbol
	DropClusterRole                                       *CommandSymbol
	SetClusterRole                                        *CommandSymbol
	AddDatabaseRole                                       *CommandSymbol
	DropDatabaseRole                                      *CommandSymbol
	SetDatabaseRole                                       *CommandSymbol
	AddTableRole                                          *CommandSymbol
	DropTableRole                                         *CommandSymbol
	SetTableRole                                          *CommandSymbol
	AddFunctionRole                                       *CommandSymbol
	DropFunctionRole                                      *CommandSymbol
	SetFunctionRole                                       *CommandSymbol
	ShowClusterBlockedPrincipals                          *CommandSymbol
	AddClusterBlockedPrincipals                           *CommandSymbol
	DropClusterBlockedPrincipals                          *CommandSymbol
	SetClusterMaintenanceMode                             *CommandSymbol
	IngestIntoTable                                       *CommandSymbol
	IngestInlineIntoTable                                 *CommandSymbol
	SetTable                                              *CommandSymbol
	AppendTable                                           *CommandSymbol
	SetOrAppendTable                                      *CommandSymbol
	SetOrReplaceTable                                     *CommandSymbol
	ExportToStorage                                       *CommandSymbol
	ExportToSqlTable                                      *CommandSymbol
	ExportToExternalTable                                 *CommandSymbol
	CreateOrAlterContinuousExport                         *CommandSymbol
	ShowContinuousExport                                  *CommandSymbol
	ShowContinuousExports                                 *CommandSymbol
	ShowClusterPendingContinuousExports                   *CommandSymbol
	ShowContinuousExportExportedArtifacts                 *CommandSymbol
	ShowContinuousExportFailures                          *CommandSymbol
	SetContinuousExportCursor                             *CommandSymbol
	DropContinuousExport                                  *CommandSymbol
	EnableContinuousExport                                *CommandSymbol
	DisableContinuousExport                               *CommandSymbol
	CreateMaterializedView                                *CommandSymbol
	CreateMaterializedViewOverMaterializedView            *CommandSymbol
	RenameMaterializedView                                *CommandSymbol
	ShowMaterializedView                                  *CommandSymbol
	ShowMaterializedViews                                 *CommandSymbol
	ShowMaterializedViewsDetails                          *CommandSymbol
	ShowMaterializedViewDetails                           *CommandSymbol
	ShowMaterializedViewPolicyRetention                   *CommandSymbol
	ShowMaterializedViewPolicyMerge                       *CommandSymbol
	ShowMaterializedViewPolicyPartitioning                *CommandSymbol
	ShowMaterializedViewExtents                           *CommandSymbol
	AlterMaterializedView                                 *CommandSymbol
	AlterMaterializedViewOverMaterializedView             *CommandSymbol
	CreateOrAlterMaterializedView                         *CommandSymbol
	CreateOrAlterMaterializedViewOverMaterializedView     *CommandSymbol
	DropMaterializedView                                  *CommandSymbol
	EnableDisableMaterializedView                         *CommandSymbol
	ShowMaterializedViewPrincipals                        *CommandSymbol
	ShowMaterializedViewSchemaAsJson                      *CommandSymbol
	ShowMaterializedViewCslSchema                         *CommandSymbol
	AlterMaterializedViewFolder                           *CommandSymbol
	AlterMaterializedViewDocString                        *CommandSymbol
	AlterMaterializedViewLookback                         *CommandSymbol
	AlterMaterializedViewAutoUpdateSchema                 *CommandSymbol
	ClearMaterializedViewData                             *CommandSymbol
	SetMaterializedViewCursor                             *CommandSymbol
	ShowTableOperationsMirroringStatus                    *CommandSymbol
	ShowTableOperationsMirroringExportedArtifacts         *CommandSymbol
	ShowTableOperationsMirroringFailures                  *CommandSymbol
	ShowCluster                                           *CommandSymbol
	ShowClusterDetails                                    *CommandSymbol
	ShowDiagnostics                                       *CommandSymbol
	ShowCapacity                                          *CommandSymbol
	ShowOperations                                        *CommandSymbol
	ShowOperationDetails                                  *CommandSymbol
	ShowJournal                                           *CommandSymbol
	ShowDatabaseJournal                                   *CommandSymbol
	ShowClusterJournal                                    *CommandSymbol
	ShowQueries                                           *CommandSymbol
	ShowRunningQueries                                    *CommandSymbol
	CancelQuery                                           *CommandSymbol
	ShowQueryPlan                                         *CommandSymbol
	ShowCache                                             *CommandSymbol
	AlterCache                                            *CommandSymbol
	ShowCommands                                          *CommandSymbol
	ShowCommandsAndQueries                                *CommandSymbol
	ShowIngestionFailures                                 *CommandSymbol
	ShowDataOperations                                    *CommandSymbol
	ShowDatabaseKeyVaultSecrets                           *CommandSymbol
	ShowClusterExtents                                    *CommandSymbol
	ShowClusterExtentsMetadata                            *CommandSymbol
	ShowDatabaseExtents                                   *CommandSymbol
	ShowDatabaseExtentsMetadata                           *CommandSymbol
	ShowDatabaseExtentTagsStatistics                      *CommandSymbol
	ShowDatabaseExtentsPartitioningStatistics             *CommandSymbol
	ShowTableExtents                                      *CommandSymbol
	ShowTableExtentsMetadata                              *CommandSymbol
	TableShardsGroupShow                                  *CommandSymbol
	TableShardsGroupMetadataShow                          *CommandSymbol
	TableShardGroupsShow                                  *CommandSymbol
	TableShardGroupsStatisticsShow                        *CommandSymbol
	TablesShardGroupsStatisticsShow                       *CommandSymbol
	DatabaseShardGroupsStatisticsShow                     *CommandSymbol
	MergeExtents                                          *CommandSymbol
	MergeExtentsDryrun                                    *CommandSymbol
	MoveExtentsFrom                                       *CommandSymbol
	MoveExtentsQuery                                      *CommandSymbol
	TableShuffleExtents                                   *CommandSymbol
	TableShuffleExtentsQuery                              *CommandSymbol
	ReplaceExtents                                        *CommandSymbol
	DropExtent                                            *CommandSymbol
	DropExtents                                           *CommandSymbol
	DropExtentsPartitionMetadata                          *CommandSymbol
	DropPretendExtentsByProperties                        *CommandSymbol
	ShowVersion                                           *CommandSymbol
	ClearTableData                                        *CommandSymbol
	ClearTableCacheStreamingIngestionSchema               *CommandSymbol
	ShowStorageArtifactsCleanupState                      *CommandSymbol
	ClusterDropStorageArtifactsCleanupState               *CommandSymbol
	StoredQueryResultSet                                  *CommandSymbol
	StoredQueryResultSetOrReplace                         *CommandSymbol
	StoredQueryResultsShow                                *CommandSymbol
	StoredQueryResultShowSchema                           *CommandSymbol
	StoredQueryResultDrop                                 *CommandSymbol
	StoredQueryResultsDrop                                *CommandSymbol
	GraphModelCreateOrAlter                               *CommandSymbol
	GraphModelShow                                        *CommandSymbol
	GraphModelsShow                                       *CommandSymbol
	GraphModelDrop                                        *CommandSymbol
	SetGraphModelAdmins                                   *CommandSymbol
	AddGraphModelAdmins                                   *CommandSymbol
	DropGraphModelAdmins                                  *CommandSymbol
	GraphSnapshotMake                                     *CommandSymbol
	GraphSnapshotShow                                     *CommandSymbol
	GraphSnapshotsShow                                    *CommandSymbol
	GraphSnapshotDrop                                     *CommandSymbol
	GraphSnapshotsDrop                                    *CommandSymbol
	GraphSnapshotShowDataStatistics                       *CommandSymbol
	GraphSnapshotsShowDataStatistics                      *CommandSymbol
	GraphSnapshotShowStatistics                           *CommandSymbol
	GraphSnapshotsShowStatistics                          *CommandSymbol
	GraphSnapshotShowFailures                             *CommandSymbol
	GraphShardMake                                        *CommandSymbol
	GraphShardDrop                                        *CommandSymbol
	ShowCertificates                                      *CommandSymbol
	ShowClockDiagnostics                                  *CommandSymbol
	ShowCloudSettings                                     *CommandSymbol
	ShowCommConcurrency                                   *CommandSymbol
	ShowCommPools                                         *CommandSymbol
	ShowFabricCache                                       *CommandSymbol
	ShowFabricLocks                                       *CommandSymbol
	ShowFabricClocks                                      *CommandSymbol
	ShowFeatureFlags                                      *CommandSymbol
	ShowMemPools                                          *CommandSymbol
	ShowServicePoints                                     *CommandSymbol
	ShowTcpConnections                                    *CommandSymbol
	ShowTcpPorts                                          *CommandSymbol
	ShowThreadPools                                       *CommandSymbol
	ShowWhatsGoingOn                                      *CommandSymbol
	ExecuteDatabaseScript                                 *CommandSymbol
	ExecuteClusterScript                                  *CommandSymbol
	CreateRequestMemoryProfile                            *CommandSymbol
	CreateRequestSupport                                  *CommandSymbol
	ShowRequestSupport                                    *CommandSymbol
	ShowClusterAdminState                                 *CommandSymbol
	ClearRemoteClusterDatabaseSchema                      *CommandSymbol
	ShowClusterMonitoring                                 *CommandSymbol
	ShowClusterScaleIn                                    *CommandSymbol
	ShowClusterServices                                   *CommandSymbol
	ShowClusterNetwork                                    *CommandSymbol
	AlterClusterStorageKeys                               *CommandSymbol
	ShowClusterStorageKeysHash                            *CommandSymbol
	AlterFabricServiceAssignmentsCommand                  *CommandSymbol
	DropFabricServiceAssignmentsCommand                   *CommandSymbol
	CreateEntityGroupCommand                              *CommandSymbol
	CreateOrAlterEntityGroupCommand                       *CommandSymbol
	AlterEntityGroup                                      *CommandSymbol
	AlterMergeEntityGroup                                 *CommandSymbol
	DropEntityGroup                                       *CommandSymbol
	ShowEntityGroup                                       *CommandSymbol
	ShowEntityGroups                                      *CommandSymbol
	AlterExtentContainersAdd                              *CommandSymbol
	AlterExtentContainersDrop                             *CommandSymbol
	AlterExtentContainersRecycle                          *CommandSymbol
	AlterExtentContainersSet                              *CommandSymbol
	ShowExtentContainers                                  *CommandSymbol
	DropEmptyExtentContainers                             *CommandSymbol
	CleanDatabaseExtentContainers                         *CommandSymbol
	ShowDatabaseExtentContainersCleanOperations           *CommandSymbol
	ClearDatabaseCacheQueryResults                        *CommandSymbol
	ShowDatabaseCacheQueryResults                         *CommandSymbol
	ShowDatabasesManagementGroups                         *CommandSymbol
	AlterDatabaseStorageKeys                              *CommandSymbol
	ClearDatabaseCacheStreamingIngestionSchema            *CommandSymbol
	ClearDatabaseCacheQueryWeakConsistency                *CommandSymbol
	ShowEntitySchema                                      *CommandSymbol
	ShowExtentDetails                                     *CommandSymbol
	ShowExtentColumnStorageStats                          *CommandSymbol
	AttachExtentsIntoTableByContainer                     *CommandSymbol
	AttachExtentsIntoTableByMetadata                      *CommandSymbol
	AlterExtentTagsFromQuery                              *CommandSymbol
	AlterMergeExtentTagsFromQuery                         *CommandSymbol
	DropExtentTagsFromQuery                               *CommandSymbol
	DropExtentTagsFromTable                               *CommandSymbol
	DropExtentTagsRetention                               *CommandSymbol
	MergeDatabaseShardGroups                              *CommandSymbol
	AlterFollowerClusterConfiguration                     *CommandSymbol
	AddFollowerDatabaseAuthorizedPrincipals               *CommandSymbol
	DropFollowerDatabaseAuthorizedPrincipals              *CommandSymbol
	AlterFollowerDatabaseAuthorizedPrincipals             *CommandSymbol
	DropFollowerDatabasePolicyCaching                     *CommandSymbol
	AlterFollowerDatabaseChildEntities                    *CommandSymbol
	AlterFollowerDatabaseConfiguration                    *CommandSymbol
	DropFollowerDatabases                                 *CommandSymbol
	ShowFollowerDatabase                                  *CommandSymbol
	AlterFollowerTablesPolicyCaching                      *CommandSymbol
	DropFollowerTablesPolicyCaching                       *CommandSymbol
	ShowFreshness                                         *CommandSymbol
	ShowFunctionSchemaAsJson                              *CommandSymbol
	SetMaterializedViewAdmins                             *CommandSymbol
	AddMaterializedViewAdmins                             *CommandSymbol
	DropMaterializedViewAdmins                            *CommandSymbol
	SetMaterializedViewConcurrency                        *CommandSymbol
	ClearMaterializedViewStatistics                       *CommandSymbol
	ShowMaterializedViewStatistics                        *CommandSymbol
	ShowMaterializedViewDiagnostics                       *CommandSymbol
	ShowMaterializedViewFailures                          *CommandSymbol
	ShowMemory                                            *CommandSymbol
	CancelOperation                                       *CommandSymbol
	DisablePlugin                                         *CommandSymbol
	EnablePlugin                                          *CommandSymbol
	ShowPlugins                                           *CommandSymbol
	ShowPrincipalAccess                                   *CommandSymbol
	ShowDatabasePurgeOperation                            *CommandSymbol
	ShowQueryExecution                                    *CommandSymbol
	AlterPoliciesOfRetention                              *CommandSymbol
	DeletePoliciesOfRetention                             *CommandSymbol
	CreateRowStore                                        *CommandSymbol
	DropRowStore                                          *CommandSymbol
	ShowRowStore                                          *CommandSymbol
	ShowRowStores                                         *CommandSymbol
	ShowRowStoreTransactions                              *CommandSymbol
	ShowRowStoreSeals                                     *CommandSymbol
	ShowSchema                                            *CommandSymbol
	ShowCallStacks                                        *CommandSymbol
	ShowFileSystem                                        *CommandSymbol
	ShowRunningCallouts                                   *CommandSymbol
	ShowStreamingIngestionFailures                        *CommandSymbol
	ShowStreamingIngestionStatistics                      *CommandSymbol
	AlterTableRowStoreReferencesDropKey                   *CommandSymbol
	AlterTableRowStoreReferencesDropRowStore              *CommandSymbol
	AlterTableRowStoreReferencesDropBlockedKeys           *CommandSymbol
	AlterTableRowStoreReferencesDisableKey                *CommandSymbol
	AlterTableRowStoreReferencesDisableRowStore           *CommandSymbol
	AlterTableRowStoreReferencesDisableBlockedKeys        *CommandSymbol
	SetTableRowStoreReferences                            *CommandSymbol
	ShowTableRowStoreReferences                           *CommandSymbol
	AlterTableColumnStatistics                            *CommandSymbol
	AlterTableColumnStatisticsMethod                      *CommandSymbol
	ShowTableColumnStatitics                              *CommandSymbol
	ShowTableDimensions                                   *CommandSymbol
	DeleteTableRecords                                    *CommandSymbol
	TableDataUpdate                                       *CommandSymbol
	DeleteMaterializedViewRecords                         *CommandSymbol
	ShowTableColumnsClassification                        *CommandSymbol
	ShowTableRowStores                                    *CommandSymbol
	ShowTableRowStoreSealInfo                             *CommandSymbol
	ShowTablesColumnStatistics                            *CommandSymbol
	ShowTableDataStatistics                               *CommandSymbol
	CreateTempStorage                                     *CommandSymbol
	DropTempStorage                                       *CommandSymbol
	DropStoredQueryResultContainers                       *CommandSymbol
	DropUnusedStoredQueryResultContainers                 *CommandSymbol
	EnableDatabaseMaintenanceMode                         *CommandSymbol
	DisableDatabaseMaintenanceMode                        *CommandSymbol
	EnableDatabaseStreamingIngestionMaintenanceMode       *CommandSymbol
	DisableDatabaseStreamingIngestionMaintenanceMode      *CommandSymbol
	ShowQueryCallTree                                     *CommandSymbol
	ShowExtentCorruptedDatetime                           *CommandSymbol
	PatchExtentCorruptedDatetime                          *CommandSymbol
	ClearClusterCredStoreCache                            *CommandSymbol
	ClearClusterOutboundAccessPolicyCache                 *CommandSymbol
	ClearClusterGroupMembershipCache                      *CommandSymbol
	ClearExternalArtifactsCache                           *CommandSymbol
	ShowDatabasesEntities                                 *CommandSymbol
	ShowDatabaseEntity                                    *CommandSymbol
	ReplaceDatabaseKeyVaultSecrets                        *CommandSymbol
	All                                                   []*CommandSymbol
}

var EngineCommands = newEngineCommandsSet()

func newEngineCommandsSet() *engineCommandsSet {
	c := &engineCommandsSet{}
	c.ShowDatabase = NewCommandSymbol("ShowDatabase", "(DatabaseName: string, PersistentStorage: string, Version: string, IsCurrent: bool, DatabaseAccessMode: string, PrettyName: string, ReservedSlot1: bool, DatabaseId: guid, InTransitionTo: string)", "DatabaseShowCommand(null, (PropertyName, PropertyValue)*)")
	c.ShowDatabaseDetails = NewCommandSymbol("ShowDatabaseDetails", "(DatabaseName: string, PersistentStorage: string, Version: string, IsCurrent: bool, DatabaseAccessMode: string, PrettyName: string, AuthorizedPrincipals: string, RetentionPolicy: string, MergePolicy: string, ReservedSlot1: string, CachingPolicy: string, ShardingPolicy: string, StreamingIngestionPolicy: string, IngestionBatchingPolicy: string, TotalSize: real, DatabaseId: guid, InTransitionTo: string)", "DatabaseShowCommand(flavor, (PropertyName, PropertyValue)*)")
	c.ShowDatabaseIdentity = NewCommandSymbol("ShowDatabaseIdentity", "(DatabaseName: string, PersistentStorage: string, Version: string, IsCurrent: bool, DatabaseAccessMode: string, PrettyName: string, CurrentUserIsUnrestrictedViewer: bool, DatabaseId: guid, InTransitionTo: string)", "DatabaseShowCommand('identity', (PropertyName, PropertyValue)*)")
	c.ShowDatabasePolicies = NewCommandSymbol("ShowDatabasePolicies", "(DatabaseName: string, PersistentStorage: string, Version: string, IsCurrent: bool, DatabaseAccessMode: string, PrettyName: string, DatabaseId: guid, AuthorizedPrincipals: string, RetentionPolicy: string, MergePolicy: string, CachingPolicy: string, ShardingPolicy: string, StreamingIngestionPolicy: string, IngestionBatchingPolicy: string)", "DatabaseShowCommand('policies', (PropertyName, PropertyValue)*)")
	c.ShowDatabaseDataStats = NewCommandSymbol("ShowDatabaseDataStats", "(DatabaseName: string, PersistentStorage: string, Version: string, IsCurrent: bool, DatabaseAccessMode: string, PrettyName: string, CurrentUseIsUnrestrictedViewer: bool, DatabaseId: guid, OriginalSize: real, ExtentSize: real, CompressedSize: real, IndexSize: real, RowCount: long, HotOriginalSize: real, HotExtentSize: real, HotCompressedSize: real, HotIndexSize: real, HotRowCount: long)", "DatabaseShowCommand('datastats', (PropertyName, PropertyValue)*)")
	c.ShowDatabaseMetadata = NewCommandSymbol("ShowDatabaseMetadata", "(DatabaseName: string, Version: string, PersistentStorage: string, MetadataPersistentStorage: string, DatabaseAccessMode: string)", "DatabaseShowCommand('metadata', (PropertyName, PropertyValue)*)")
	c.ShowClusterDatabases = NewCommandSymbol("ShowClusterDatabases", "(DatabaseName: string, PersistentStorage: string, Version: string, IsCurrent: bool, DatabaseAccessMode: string, PrettyName: string, ReservedSlot1: bool, DatabaseId: guid, InTransitionTo: string)", "DatabasesShowCommand(null, DatabaseName*, (PropertyName, PropertyValue)*)")
	c.ShowClusterDatabasesDetails = NewCommandSymbol("ShowClusterDatabasesDetails", "(DatabaseName: string, PersistentStorage: string, Version: string, IsCurrent: bool, DatabaseAccessMode: string, PrettyName: string, AuthorizedPrincipals: string, RetentionPolicy: string, MergePolicy: string, ReservedSlot1: string, CachingPolicy: string, ShardingPolicy: string, StreamingIngestionPolicy: string, IngestionBatchingPolicy: string, TotalSize: real, DatabaseId: guid, InTransitionTo: string)", "DatabasesShowCommand('details', DatabaseName*, (PropertyName, PropertyValue)*)")
	c.ShowClusterDatabasesIdentity = NewCommandSymbol("ShowClusterDatabasesIdentity", "(DatabaseName: string, PersistentStorage: string, Version: string, IsCurrent: bool, DatabaseAccessMode: string, PrettyName: string, CurrentUserIsUnrestrictedViewer: bool, DatabaseId: guid, InTransitionTo: string)", "DatabasesShowCommand('identity', DatabaseName*, (PropertyName, PropertyValue)*)")
	c.ShowClusterDatabasesPolicies = NewCommandSymbol("ShowClusterDatabasesPolicies", "(DatabaseName: string, PersistentStorage: string, Version: string, IsCurrent: bool, DatabaseAccessMode: string, PrettyName: string, DatabaseId: guid, AuthorizedPrincipals: string, RetentionPolicy: string, MergePolicy: string, CachingPolicy: string, ShardingPolicy: string, StreamingIngestionPolicy: string, IngestionBatchingPolicy: string)", "DatabasesShowCommand('policies', DatabaseName*, (PropertyName, PropertyValue)*)")
	c.ShowClusterDatabasesDataStats = NewCommandSymbol("ShowClusterDatabasesDataStats", "(DatabaseName: string, PersistentStorage: string, Version: string, IsCurrent: bool, DatabaseAccessMode: string, PrettyName: string, DatabaseId: guid, OriginalSize: real, ExtentSize: real, CompressedSize: real, IndexSize: real, RowCount: long, HotOriginalSize: real, HotExtentSize: real, HotCompressedSize: real, HotIndexSize: real, HotRowCount: long)", "DatabasesShowCommand('datastats', DatabaseName*, (PropertyName, PropertyValue)*)")
	c.ShowClusterDatabasesMetadata = NewCommandSymbol("ShowClusterDatabasesMetadata", "(DatabaseName: string, Version: string, PersistentStorage: string, MetadataPersistentStorage: string, DatabaseAccessMode: string)", "DatabasesShowCommand('metadata', DatabaseName*, (PropertyName, PropertyValue)*)")
	c.CreateDatabase = NewCommandSymbol("CreateDatabase", "(DatabaseName: string, PersistentPath: string, Created: string, StoresMetadata: bool, StoresData: bool)", "DatabaseCreateCommand(DatabaseName, Path*, IfNotExists?, (PropertyName, PropertyValue)*)")
	c.AttachDatabase = NewCommandSymbol("AttachDatabase", "(Step: string, Duration: string)", "DatabaseAttachCommand(DatabaseName, Path, ReadOnly ?? 'ReadWrite', Version, (PropertyName, PropertyValue)*)")
	c.DetachDatabase = NewCommandSymbol("DetachDatabase", "(Table: string, NumberOfRemovedExtents: string)", "DatabaseDetachCommand(DatabaseName, IfExists?, SkipSeal?)")
	c.AlterDatabasePrettyName = NewCommandSymbol("AlterDatabasePrettyName", "(DatabaseName: string, PrettyName: string)", "DatabaseSetPrettyNameCommand(DatabaseName, PrettyName)")
	c.DropDatabasePrettyName = NewCommandSymbol("DropDatabasePrettyName", "(DatabaseName: string, PrettyName: string)", "DatabaseSetPrettyNameCommand(DatabaseName, null)")
	c.AlterDatabasePersistMetadata = NewCommandSymbol("AlterDatabasePersistMetadata", "(Moniker: guid, Url: string, State: string, CreatedOn: datetime, MaxDateTime: datetime, IsRecyclable: bool, StoresDatabaseMetadata: bool, HardDeletePeriod: timespan)", "DatabaseMetadataContainerAlterCommand(DatabaseName, Path, AllowNonEmptyContainer?)")
	c.SetAccess = NewCommandSymbol("SetAccess", "(DatabaseName: string, RequestedAccessMode: string, Status: string)", "DatabaseSetAccessModeCommand(DatabaseName, AccessMode)")
	c.ShowDatabaseSchema = NewCommandSymbol("ShowDatabaseSchema", "(DatabaseName: string, TableName: string, ColumnName: string, ColumnType: string, IsDefaultTable: bool, IsDefaultColumn: bool, PrettyName: string, Version: string, Folder: string, DocString: string)", "SchemaShowCommand(Details?, (DatabaseName, Version))")
	c.ShowDatabaseSchemaAsJson = NewCommandSymbol("ShowDatabaseSchemaAsJson", "(DatabaseSchema: string)", "SchemaShowAsJsonCommand((DatabaseName, Version), (PropertyName, PropertyValue)*)")
	c.ShowDatabaseSchemaAsCslScript = NewCommandSymbol("ShowDatabaseSchemaAsCslScript", "(DatabaseSchemaScript: string)", "DatabaseSchemaShowAsCslCommand(Script?, (DatabaseName, Version), (PropertyName, PropertyValue)*)")
	c.ShowDatabaseCslSchema = NewCommandSymbol("ShowDatabaseCslSchema", "(TableName: string, Schema: string, DatabaseName: string, Folder: string, DocString: string)", "DatabaseSchemaShowAsCslCommand(Script?, (DatabaseName, Version))")
	c.ShowDatabaseSchemaViolations = NewCommandSymbol("ShowDatabaseSchemaViolations", "(EntityKind: string, EntityName: string, Property: string, Reason: string)", "DatabaseSchemaViolationsShowCommand(DatabaseName)")
	c.ShowDatabasesSchema = NewCommandSymbol("ShowDatabasesSchema", "(DatabaseName: string, TableName: string, ColumnName: string, ColumnType: string, IsDefaultTable: bool, IsDefaultColumn: bool, PrettyName: string, Version: string, Folder: string, DocString: string)", "SchemaShowCommand(Details?, (DatabaseName, Version)*)")
	c.ShowDatabasesSchemaAsJson = NewCommandSymbol("ShowDatabasesSchemaAsJson", "(DatabaseSchema: string)", "SchemaShowAsJsonCommand((DatabaseName, Version)*, (PropertyName, PropertyValue)*)")
	c.CreateDatabaseIngestionMapping = NewCommandSymbol("CreateDatabaseIngestionMapping", "(Name: string, Kind: string, Mapping: string, LastUpdatedOn: datetime, Database: string)")
	c.CreateOrAlterDatabaseIngestionMapping = NewCommandSymbol("CreateOrAlterDatabaseIngestionMapping", "(Name: string, Kind: string, Mapping: string, LastUpdatedOn: datetime, Database: string)")
	c.AlterDatabaseIngestionMapping = NewCommandSymbol("AlterDatabaseIngestionMapping", "(Name: string, Kind: string, Mapping: string, LastUpdatedOn: datetime, Database: string)")
	c.AlterMergeDatabaseIngestionMapping = NewCommandSymbol("AlterMergeDatabaseIngestionMapping", "(Name: string, Kind: string, Mapping: string, LastUpdatedOn: datetime, Database: string)")
	c.ShowDatabaseIngestionMappings = NewCommandSymbol("ShowDatabaseIngestionMappings", "(Name: string, Kind: string, Mapping: string, LastUpdatedOn: datetime, Database: string)")
	c.ShowIngestionMappings = NewCommandSymbol("ShowIngestionMappings", "()")
	c.DropDatabaseIngestionMapping = NewCommandSymbol("DropDatabaseIngestionMapping", "(Name: string, Kind: string, Mapping: string, LastUpdatedOn: datetime, Database: string)")
	c.ShowTables = NewCommandSymbol("ShowTables", "(TableName: string, DatabaseName: string, Folder: string, DocString: string)")
	c.ShowTable = NewCommandSymbol("ShowTable", "(AttributeName: string, AttributeType: string, ExtentSize: long, CompressionRatio: real, IndexSize: long, IndexSizePercent: real, OriginalSize: long, AttributeId: guid, SharedIndexSize: long, StorageEngineVersion: string)")
	c.ShowTablesDetails = NewCommandSymbol("ShowTablesDetails", "(TableName: string, DatabaseName: string, Folder: string, DocString: string, TotalExtents: long, TotalExtentSize: real, TotalOriginalSize: real, TotalRowCount: long, HotExtents: long, HotExtentSize: real, HotOriginalSize: real, HotRowCount: long, AuthorizedPrincipals: string, RetentionPolicy: string, CachingPolicy: string, ShardingPolicy: string, MergePolicy: string, StreamingIngestionPolicy: string, IngestionBatchingPolicy: string, MinExtentsCreationTime: datetime, MaxExtentsCreationTime: datetime, RowOrderPolicy: string, TableId: guid, RowLevelSecurityPolicy: string, AutoDeletePolicy: string, UpdatePolicy: string, UpdatePolicyStates: string)")
	c.ShowTableDetails = NewCommandSymbol("ShowTableDetails", "(TableName: string, DatabaseName: string, Folder: string, DocString: string, TotalExtents: long, TotalExtentSize: real, TotalOriginalSize: real, TotalRowCount: long, HotExtents: long, HotExtentSize: real, HotOriginalSize: real, HotRowCount: long, AuthorizedPrincipals: string, RetentionPolicy: string, CachingPolicy: string, ShardingPolicy: string, MergePolicy: string, StreamingIngestionPolicy: string, IngestionBatchingPolicy: string, MinExtentsCreationTime: datetime, MaxExtentsCreationTime: datetime, RowOrderPolicy: string, TableId: guid, RowLevelSecurityPolicy: string, AutoDeletePolicy: string, UpdatePolicy: string, UpdatePolicyStates: string)")
	c.ShowTableCslSchema = NewCommandSymbol("ShowTableCslSchema", "(TableName: string, Schema: string, DatabaseName: string, Folder: string, DocString: string)")
	c.ShowTableSchemaAsJson = NewCommandSymbol("ShowTableSchemaAsJson", "(TableName: string, Schema: string, DatabaseName: string, Folder: string, DocString: string)")
	c.CreateTable = NewCommandSymbol("CreateTable", "(TableName: string, Schema: string, DatabaseName: string, Folder: string, DocString: string)")
	c.CreateTableBasedOnAnother = NewCommandSymbol("CreateTableBasedOnAnother", "(TableName: string, Schema: string, DatabaseName: string, Folder: string, DocString: string)")
	c.CreateMergeTable = NewCommandSymbol("CreateMergeTable", "(TableName: string, Schema: string, DatabaseName: string, Folder: string, DocString: string)")
	c.DefineTable = NewCommandSymbol("DefineTable", "(TableName: string, Schema: string, DatabaseName: string, Folder: string, DocString: string)")
	c.CreateTables = NewCommandSymbol("CreateTables", "(TableName: string, DatabaseName: string, Folder: string, DocString: string)")
	c.CreateMergeTables = NewCommandSymbol("CreateMergeTables", "(TableName: string, DatabaseName: string, Folder: string, DocString: string)")
	c.DefineTables = NewCommandSymbol("DefineTables", "(TableName: string, DatabaseName: string, Folder: string, DocString: string)")
	c.AlterTable = NewCommandSymbol("AlterTable", "(TableName: string, Schema: string, DatabaseName: string, Folder: string, DocString: string)")
	c.AlterMergeTable = NewCommandSymbol("AlterMergeTable", "(TableName: string, Schema: string, DatabaseName: string, Folder: string, DocString: string)")
	c.AlterTableDocString = NewCommandSymbol("AlterTableDocString", "(TableName: string, Schema: string, DatabaseName: string, Folder: string, DocString: string)")
	c.AlterTableFolder = NewCommandSymbol("AlterTableFolder", "(TableName: string, Schema: string, DatabaseName: string, Folder: string, DocString: string)")
	c.RenameTable = NewCommandSymbol("RenameTable", "(TableName: string, DatabaseName: string, Folder: string, DocString: string)")
	c.RenameTables = NewCommandSymbol("RenameTables", "(TableName: string, DatabaseName: string, Folder: string, DocString: string)")
	c.UndoDropExtentContainer = NewCommandSymbol("UndoDropExtentContainer", "(ContainerId: guid, RestoreContainerFinalState: string)")
	c.DropTable = NewCommandSymbol("DropTable", "(TableName: string, DatabaseName: string)")
	c.UndoDropTable = NewCommandSymbol("UndoDropTable", "(ExtentId: guid, NumberOfRecords: long, Status: string, FailureReason: string)")
	c.DropTables = NewCommandSymbol("DropTables", "(TableName: string, DatabaseName: string)")
	c.CreateTableIngestionMapping = NewCommandSymbol("CreateTableIngestionMapping", "(Name: string, Kind: string, Mapping: string, LastUpdatedOn: datetime, Database: string, Table: string)")
	c.CreateOrAlterTableIngestionMapping = NewCommandSymbol("CreateOrAlterTableIngestionMapping", "(Name: string, Kind: string, Mapping: string, LastUpdatedOn: datetime, Database: string, Table: string)")
	c.AlterTableIngestionMapping = NewCommandSymbol("AlterTableIngestionMapping", "(Name: string, Kind: string, Mapping: string, LastUpdatedOn: datetime, Database: string, Table: string)")
	c.AlterMergeTableIngestionMapping = NewCommandSymbol("AlterMergeTableIngestionMapping", "(Name: string, Kind: string, Mapping: string, LastUpdatedOn: datetime, Database: string, Table: string)")
	c.ShowTableIngestionMappings = NewCommandSymbol("ShowTableIngestionMappings", "(Name: string, Kind: string, Mapping: string, LastUpdatedOn: datetime, Database: string, Table: string)")
	c.ShowTableIngestionMapping = NewCommandSymbol("ShowTableIngestionMapping", "(Name: string, Kind: string, Mapping: string, LastUpdatedOn: datetime, Database: string, Table: string)")
	c.DropTableIngestionMapping = NewCommandSymbol("DropTableIngestionMapping", "(Name: string, Kind: string, Mapping: string, LastUpdatedOn: datetime, Database: string, Table: string)")
	c.RenameColumn = NewCommandSymbol("RenameColumn", "(EntityName: string, DataType: string, Policy: string)")
	c.RenameColumns = NewCommandSymbol("RenameColumns", "(EntityName: string, DataType: string, Policy: string)")
	c.AlterColumnType = NewCommandSymbol("AlterColumnType", "(EntityName: string, DataType: string, Policy: string)")
	c.DropColumn = NewCommandSymbol("DropColumn", "(TableName: string, Schema: string, DatabaseName: string, Folder: string, DocString: string)")
	c.DropTableColumns = NewCommandSymbol("DropTableColumns", "(TableName: string, Schema: string, DatabaseName: string, Folder: string, DocString: string)")
	c.AlterTableColumnDocStrings = NewCommandSymbol("AlterTableColumnDocStrings", "(TableName: string, Schema: string, DatabaseName: string, Folder: string, DocString: string)")
	c.AlterMergeTableColumnDocStrings = NewCommandSymbol("AlterMergeTableColumnDocStrings", "(TableName: string, Schema: string, DatabaseName: string, Folder: string, DocString: string)")
	c.ShowFunctions = NewCommandSymbol("ShowFunctions", "(Name: string, Parameters: string, Body: string, Folder: string, DocString: string)")
	c.ShowFunction = NewCommandSymbol("ShowFunction", "(Name: string, Parameters: string, Body: string, Folder: string, DocString: string)")
	c.CreateFunction = NewCommandSymbol("CreateFunction", "(Name: string, Parameters: string, Body: string, Folder: string, DocString: string)")
	c.AlterFunction = NewCommandSymbol("AlterFunction", "(Name: string, Parameters: string, Body: string, Folder: string, DocString: string)")
	c.CreateOrAlterFunction = NewCommandSymbol("CreateOrAlterFunction", "(Name: string, Parameters: string, Body: string, Folder: string, DocString: string)")
	c.DropFunction = NewCommandSymbol("DropFunction", "(Name: string)")
	c.DropFunctions = NewCommandSymbol("DropFunctions", "(Name: string, Parameters: string, Body: string, Folder: string, DocString: string)")
	c.AlterFunctionDocString = NewCommandSymbol("AlterFunctionDocString", "(Name: string, Parameters: string, Body: string, Folder: string, DocString: string)")
	c.AlterFunctionFolder = NewCommandSymbol("AlterFunctionFolder", "(Name: string, Parameters: string, Body: string, Folder: string, DocString: string)")
	c.ShowExternalTables = NewCommandSymbol("ShowExternalTables", "(TableName: string, TableType: string, Folder: string, DocString: string, Properties: string, ConnectionStrings: dynamic, Partitions: dynamic, PathFormat: string)")
	c.ShowExternalTable = NewCommandSymbol("ShowExternalTable", "(TableName: string, TableType: string, Folder: string, DocString: string, Properties: string, ConnectionStrings: dynamic, Partitions: dynamic, PathFormat: string)")
	c.ShowExternalTablesDetails = NewCommandSymbol("ShowExternalTablesDetails", "(TableName: string, QueryAccelerationPolicy: string, QueryAccelerationState: string)")
	c.ShowExternalTableDetails = NewCommandSymbol("ShowExternalTableDetails", "(TableName: string, QueryAccelerationPolicy: string, QueryAccelerationState: string)")
	c.ShowExternalTableCslSchema = NewCommandSymbol("ShowExternalTableCslSchema", "(TableName: string, Schema: string, DatabaseName: string, Folder: string, DocString: string)")
	c.ShowExternalTableSchema = NewCommandSymbol("ShowExternalTableSchema", "(TableName: string, Schema: string, DatabaseName: string, Folder: string, DocString: string)")
	c.ShowExternalTableArtifacts = NewCommandSymbol("ShowExternalTableArtifacts", "(Uri: string, Partition: dynamic, Size: long)")
	c.DropExternalTable = NewCommandSymbol("DropExternalTable", "(TableName: string, TableType: string, Folder: string, DocString: string, Properties: string, ConnectionStrings: dynamic, Partitions: dynamic, PathFormat: string)")
	c.CreateStorageExternalTable = NewCommandSymbol("CreateStorageExternalTable", "(TableName: string, TableType: string, Folder: string, DocString: string, Properties: string, ConnectionStrings: dynamic, Partitions: dynamic, PathFormat: string)")
	c.AlterStorageExternalTable = NewCommandSymbol("AlterStorageExternalTable", "(TableName: string, TableType: string, Folder: string, DocString: string, Schema: string, Properties: string)")
	c.CreateOrAlterStorageExternalTable = NewCommandSymbol("CreateOrAlterStorageExternalTable", "(TableName: string, TableType: string, Folder: string, DocString: string, Schema: string, Properties: string)")
	c.CreateSqlExternalTable = NewCommandSymbol("CreateSqlExternalTable", "(TableName: string, TableType: string, Folder: string, DocString: string, Properties: string, ConnectionStrings: dynamic, Partitions: dynamic, PathFormat: string)")
	c.AlterSqlExternalTable = NewCommandSymbol("AlterSqlExternalTable", "(TableName: string, TableType: string, Folder: string, DocString: string, Schema: string, Properties: string)")
	c.CreateOrAlterSqlExternalTable = NewCommandSymbol("CreateOrAlterSqlExternalTable", "(TableName: string, TableType: string, Folder: string, DocString: string, Schema: string, Properties: string)")
	c.CreateExternalTableMapping = NewCommandSymbol("CreateExternalTableMapping", "(Name: string, Kind: string, Mapping: string, LastUpdatedOn: datetime, Database: string, Table: string)")
	c.SetExternalTableAdmins = NewCommandSymbol("SetExternalTableAdmins", "()")
	c.AddExternalTableAdmins = NewCommandSymbol("AddExternalTableAdmins", "()")
	c.DropExternalTableAdmins = NewCommandSymbol("DropExternalTableAdmins", "()")
	c.AlterExternalTableDocString = NewCommandSymbol("AlterExternalTableDocString", "()")
	c.AlterExternalTableFolder = NewCommandSymbol("AlterExternalTableFolder", "()")
	c.ShowExternalTablePrincipals = NewCommandSymbol("ShowExternalTablePrincipals", "()")
	c.ShowFabric = NewCommandSymbol("ShowFabric", "()")
	c.AlterExternalTableMapping = NewCommandSymbol("AlterExternalTableMapping", "(Name: string, Kind: string, Mapping: string, LastUpdatedOn: datetime, Database: string, Table: string)")
	c.ShowExternalTableMappings = NewCommandSymbol("ShowExternalTableMappings", "(Name: string, Kind: string, Mapping: string, LastUpdatedOn: datetime, Database: string, Table: string)")
	c.ShowExternalTableMapping = NewCommandSymbol("ShowExternalTableMapping", "(Name: string, Kind: string, Mapping: string, LastUpdatedOn: datetime, Database: string, Table: string)")
	c.DropExternalTableMapping = NewCommandSymbol("DropExternalTableMapping", "(Name: string, Kind: string, Mapping: string, LastUpdatedOn: datetime, Database: string, Table: string)")
	c.ShowWorkloadGroups = NewCommandSymbol("ShowWorkloadGroups", "(WorkloadGroupName: string, WorkloadGroup:string)")
	c.ShowWorkloadGroup = NewCommandSymbol("ShowWorkloadGroup", "(WorkloadGroupName: string, WorkloadGroup:string)")
	c.CreateOrAleterWorkloadGroup = NewCommandSymbol("CreateOrAleterWorkloadGroup", "(WorkloadGroupName: string, WorkloadGroup:string)")
	c.AlterMergeWorkloadGroup = NewCommandSymbol("AlterMergeWorkloadGroup", "(WorkloadGroupName: string, WorkloadGroup:string)")
	c.DropWorkloadGroup = NewCommandSymbol("DropWorkloadGroup", "(WorkloadGroupName: string, WorkloadGroup:string)")
	c.ShowDatabasePolicyCaching = NewCommandSymbol("ShowDatabasePolicyCaching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTablePolicyCaching = NewCommandSymbol("ShowTablePolicyCaching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTableStarPolicyCaching = NewCommandSymbol("ShowTableStarPolicyCaching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowColumnPolicyCaching = NewCommandSymbol("ShowColumnPolicyCaching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowMaterializedViewPolicyCaching = NewCommandSymbol("ShowMaterializedViewPolicyCaching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowGraphModelPolicyCaching = NewCommandSymbol("ShowGraphModelPolicyCaching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowGraphModelStarPolicyCaching = NewCommandSymbol("ShowGraphModelStarPolicyCaching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowClusterPolicyCaching = NewCommandSymbol("ShowClusterPolicyCaching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterDatabasePolicyCaching = NewCommandSymbol("AlterDatabasePolicyCaching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablePolicyCaching = NewCommandSymbol("AlterTablePolicyCaching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablesPolicyCaching = NewCommandSymbol("AlterTablesPolicyCaching", "()")
	c.AlterColumnPolicyCaching = NewCommandSymbol("AlterColumnPolicyCaching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMaterializedViewPolicyCaching = NewCommandSymbol("AlterMaterializedViewPolicyCaching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterGraphModelPolicyCaching = NewCommandSymbol("AlterGraphModelPolicyCaching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterClusterPolicyCaching = NewCommandSymbol("AlterClusterPolicyCaching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteDatabasePolicyCaching = NewCommandSymbol("DeleteDatabasePolicyCaching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteTablePolicyCaching = NewCommandSymbol("DeleteTablePolicyCaching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteColumnPolicyCaching = NewCommandSymbol("DeleteColumnPolicyCaching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteMaterializedViewPolicyCaching = NewCommandSymbol("DeleteMaterializedViewPolicyCaching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteGraphModelPolicyCaching = NewCommandSymbol("DeleteGraphModelPolicyCaching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteClusterPolicyCaching = NewCommandSymbol("DeleteClusterPolicyCaching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTablePolicyChangeDataCapture = NewCommandSymbol("ShowTablePolicyChangeDataCapture", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablePolicyChangeDataCapture = NewCommandSymbol("AlterTablePolicyChangeDataCapture", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablesPolicyChangeDataCapture = NewCommandSymbol("AlterTablesPolicyChangeDataCapture", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteTablePolicyChangeDataCapture = NewCommandSymbol("DeleteTablePolicyChangeDataCapture", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTablePolicyIngestionTime = NewCommandSymbol("ShowTablePolicyIngestionTime", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTableStarPolicyIngestionTime = NewCommandSymbol("ShowTableStarPolicyIngestionTime", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablePolicyIngestionTime = NewCommandSymbol("AlterTablePolicyIngestionTime", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablesPolicyIngestionTime = NewCommandSymbol("AlterTablesPolicyIngestionTime", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteTablePolicyIngestionTime = NewCommandSymbol("DeleteTablePolicyIngestionTime", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowDatabasePolicyRowId = NewCommandSymbol("ShowDatabasePolicyRowId", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterDatabasePolicyRowId = NewCommandSymbol("AlterDatabasePolicyRowId", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteDatabasePolicyRowId = NewCommandSymbol("DeleteDatabasePolicyRowId", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTablePolicyRowId = NewCommandSymbol("ShowTablePolicyRowId", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablePolicyRowId = NewCommandSymbol("AlterTablePolicyRowId", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablesPolicyRowId = NewCommandSymbol("AlterTablesPolicyRowId", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteTablePolicyRowId = NewCommandSymbol("DeleteTablePolicyRowId", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTablePolicyRetention = NewCommandSymbol("ShowTablePolicyRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTableStarPolicyRetention = NewCommandSymbol("ShowTableStarPolicyRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowGraphPolicyRetention = NewCommandSymbol("ShowGraphPolicyRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowGraphStarPolicyRetention = NewCommandSymbol("ShowGraphStarPolicyRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowDatabasePolicyRetention = NewCommandSymbol("ShowDatabasePolicyRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablePolicyRetention = NewCommandSymbol("AlterTablePolicyRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMaterializedViewPolicyRetention = NewCommandSymbol("AlterMaterializedViewPolicyRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterDatabasePolicyRetention = NewCommandSymbol("AlterDatabasePolicyRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterGraphModelPolicyRetention = NewCommandSymbol("AlterGraphModelPolicyRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablesPolicyRetention = NewCommandSymbol("AlterTablesPolicyRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeTablePolicyRetention = NewCommandSymbol("AlterMergeTablePolicyRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeMaterializedViewPolicyRetention = NewCommandSymbol("AlterMergeMaterializedViewPolicyRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeDatabasePolicyRetention = NewCommandSymbol("AlterMergeDatabasePolicyRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteTablePolicyRetention = NewCommandSymbol("DeleteTablePolicyRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteDatabasePolicyRetention = NewCommandSymbol("DeleteDatabasePolicyRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowDatabasePolicyHardRetentionViolations = NewCommandSymbol("ShowDatabasePolicyHardRetentionViolations", "()")
	c.ShowDatabasePolicySoftRetentionViolations = NewCommandSymbol("ShowDatabasePolicySoftRetentionViolations", "()")
	c.ShowTablePolicyRowLevelSecurity = NewCommandSymbol("ShowTablePolicyRowLevelSecurity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTableStarPolicyRowLevelSecurity = NewCommandSymbol("ShowTableStarPolicyRowLevelSecurity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablePolicyRowLevelSecurity = NewCommandSymbol("AlterTablePolicyRowLevelSecurity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteTablePolicyRowLevelSecurity = NewCommandSymbol("DeleteTablePolicyRowLevelSecurity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowMaterializedViewPolicyRowLevelSecurity = NewCommandSymbol("ShowMaterializedViewPolicyRowLevelSecurity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMaterializedViewPolicyRowLevelSecurity = NewCommandSymbol("AlterMaterializedViewPolicyRowLevelSecurity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteMaterializedViewPolicyRowLevelSecurity = NewCommandSymbol("DeleteMaterializedViewPolicyRowLevelSecurity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowExternalTablePolicyRowLevelSecurity = NewCommandSymbol("ShowExternalTablePolicyRowLevelSecurity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowExternalTableStarPolicyRowLevelSecurity = NewCommandSymbol("ShowExternalTableStarPolicyRowLevelSecurity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterExternalTablePolicyRowLevelSecurity = NewCommandSymbol("AlterExternalTablePolicyRowLevelSecurity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteExternalTablePolicyRowLevelSecurity = NewCommandSymbol("DeleteExternalTablePolicyRowLevelSecurity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTablePolicyRowOrder = NewCommandSymbol("ShowTablePolicyRowOrder", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTableStarPolicyRowOrder = NewCommandSymbol("ShowTableStarPolicyRowOrder", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablePolicyRowOrder = NewCommandSymbol("AlterTablePolicyRowOrder", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablesPolicyRowOrder = NewCommandSymbol("AlterTablesPolicyRowOrder", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeTablePolicyRowOrder = NewCommandSymbol("AlterMergeTablePolicyRowOrder", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteTablePolicyRowOrder = NewCommandSymbol("DeleteTablePolicyRowOrder", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTablePolicyUpdate = NewCommandSymbol("ShowTablePolicyUpdate", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTableStarPolicyUpdate = NewCommandSymbol("ShowTableStarPolicyUpdate", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablePolicyUpdate = NewCommandSymbol("AlterTablePolicyUpdate", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeTablePolicyUpdate = NewCommandSymbol("AlterMergeTablePolicyUpdate", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteTablePolicyUpdate = NewCommandSymbol("DeleteTablePolicyUpdate", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowClusterPolicyIngestionBatching = NewCommandSymbol("ShowClusterPolicyIngestionBatching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowDatabasePolicyIngestionBatching = NewCommandSymbol("ShowDatabasePolicyIngestionBatching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTablePolicyIngestionBatching = NewCommandSymbol("ShowTablePolicyIngestionBatching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTableStarPolicyIngestionBatching = NewCommandSymbol("ShowTableStarPolicyIngestionBatching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterClusterPolicyIngestionBatching = NewCommandSymbol("AlterClusterPolicyIngestionBatching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeClusterPolicyIngestionBatching = NewCommandSymbol("AlterMergeClusterPolicyIngestionBatching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterDatabasePolicyIngestionBatching = NewCommandSymbol("AlterDatabasePolicyIngestionBatching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeDatabasePolicyIngestionBatching = NewCommandSymbol("AlterMergeDatabasePolicyIngestionBatching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablePolicyIngestionBatching = NewCommandSymbol("AlterTablePolicyIngestionBatching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeTablePolicyIngestionBatching = NewCommandSymbol("AlterMergeTablePolicyIngestionBatching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablesPolicyIngestionBatching = NewCommandSymbol("AlterTablesPolicyIngestionBatching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteDatabasePolicyIngestionBatching = NewCommandSymbol("DeleteDatabasePolicyIngestionBatching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteTablePolicyIngestionBatching = NewCommandSymbol("DeleteTablePolicyIngestionBatching", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowDatabasePolicyEncoding = NewCommandSymbol("ShowDatabasePolicyEncoding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTablePolicyEncoding = NewCommandSymbol("ShowTablePolicyEncoding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowColumnPolicyEncoding = NewCommandSymbol("ShowColumnPolicyEncoding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterDatabasePolicyEncoding = NewCommandSymbol("AlterDatabasePolicyEncoding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablePolicyEncoding = NewCommandSymbol("AlterTablePolicyEncoding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTableColumnsPolicyEncoding = NewCommandSymbol("AlterTableColumnsPolicyEncoding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterColumnPolicyEncoding = NewCommandSymbol("AlterColumnPolicyEncoding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterColumnsPolicyEncodingByQuery = NewCommandSymbol("AlterColumnsPolicyEncodingByQuery", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterColumnPolicyEncodingType = NewCommandSymbol("AlterColumnPolicyEncodingType", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeDatabasePolicyEncoding = NewCommandSymbol("AlterMergeDatabasePolicyEncoding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeTablePolicyEncoding = NewCommandSymbol("AlterMergeTablePolicyEncoding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeColumnPolicyEncoding = NewCommandSymbol("AlterMergeColumnPolicyEncoding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteDatabasePolicyEncoding = NewCommandSymbol("DeleteDatabasePolicyEncoding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteTablePolicyEncoding = NewCommandSymbol("DeleteTablePolicyEncoding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteColumnPolicyEncoding = NewCommandSymbol("DeleteColumnPolicyEncoding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowDatabasePolicyMerge = NewCommandSymbol("ShowDatabasePolicyMerge", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTablePolicyMerge = NewCommandSymbol("ShowTablePolicyMerge", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTableStarPolicyMerge = NewCommandSymbol("ShowTableStarPolicyMerge", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterDatabasePolicyMerge = NewCommandSymbol("AlterDatabasePolicyMerge", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablePolicyMerge = NewCommandSymbol("AlterTablePolicyMerge", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablesPolicyMerge = NewCommandSymbol("AlterTablesPolicyMerge", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeDatabasePolicyMerge = NewCommandSymbol("AlterMergeDatabasePolicyMerge", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeTablePolicyMerge = NewCommandSymbol("AlterMergeTablePolicyMerge", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeMaterializedViewPolicyMerge = NewCommandSymbol("AlterMergeMaterializedViewPolicyMerge", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteDatabasePolicyMerge = NewCommandSymbol("DeleteDatabasePolicyMerge", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteTablePolicyMerge = NewCommandSymbol("DeleteTablePolicyMerge", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowExternalTablePolicyQueryAcceleration = NewCommandSymbol("ShowExternalTablePolicyQueryAcceleration", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowExternalTablesPolicyQueryAcceleration = NewCommandSymbol("ShowExternalTablesPolicyQueryAcceleration", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterExternalTablePolicyQueryAcceleration = NewCommandSymbol("AlterExternalTablePolicyQueryAcceleration", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeExternalTablePolicyQueryAcceleration = NewCommandSymbol("AlterMergeExternalTablePolicyQueryAcceleration", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteExternalTablePolicyQueryAcceleration = NewCommandSymbol("DeleteExternalTablePolicyQueryAcceleration", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowExternalTableQueryAccelerationStatatistics = NewCommandSymbol("ShowExternalTableQueryAccelerationStatatistics", "(ExternalTableName: string, IsEnabled: bool, Hot: timespan, HotSize: long, LastUpdatedDateTime: datetime, AccelerationPendingDataFilesCount:long, AccelerationPendingDataFilesSize:long, AccelerationCompletePercentage: double, NotHealthyReason: string)")
	c.ShowExternalTablesQueryAccelerationStatatistics = NewCommandSymbol("ShowExternalTablesQueryAccelerationStatatistics", "(ExternalTableName: string, IsEnabled: bool, Hot: timespan, HotSize: long, LastUpdatedDateTime: datetime, AccelerationPendingDataFilesCount:long, AccelerationPendingDataFilesSize:long, AccelerationCompletePercentage: double, NotHealthyReason: string)")
	c.AlterTablePolicyMirroring = NewCommandSymbol("AlterTablePolicyMirroring", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeTablePolicyMirroring = NewCommandSymbol("AlterMergeTablePolicyMirroring", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablePolicyMirroringWithJson = NewCommandSymbol("AlterTablePolicyMirroringWithJson", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeTablePolicyMirroringWithJson = NewCommandSymbol("AlterMergeTablePolicyMirroringWithJson", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteTablePolicyMirroring = NewCommandSymbol("DeleteTablePolicyMirroring", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTablePolicyMirroring = NewCommandSymbol("ShowTablePolicyMirroring", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTableStarPolicyMirroring = NewCommandSymbol("ShowTableStarPolicyMirroring", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.CreateMirroringTemplate = NewCommandSymbol("CreateMirroringTemplate", "(Name: string, Kind: string, ConnectionString: string, IsEnabled: bool, AutoApplyToNewTables: bool)")
	c.CreateOrAlterMirroringTemplate = NewCommandSymbol("CreateOrAlterMirroringTemplate", "(Name: string, Kind: string, ConnectionString: string, IsEnabled: bool, AutoApplyToNewTables: bool)")
	c.AlterMirroringTemplate = NewCommandSymbol("AlterMirroringTemplate", "(Name: string, Kind: string, ConnectionString: string, IsEnabled: bool, AutoApplyToNewTables: bool)")
	c.AlterMergeMirroringTemplate = NewCommandSymbol("AlterMergeMirroringTemplate", "(Name: string, Kind: string, ConnectionString: string, IsEnabled: bool, AutoApplyToNewTables: bool)")
	c.DeleteMirroringTemplate = NewCommandSymbol("DeleteMirroringTemplate", "(Name: string, Kind: string, ConnectionString: string, IsEnabled: bool, AutoApplyToNewTables: bool)")
	c.ShowMirroringTemplate = NewCommandSymbol("ShowMirroringTemplate", "(Name: string, Kind: string, ConnectionString: string, IsEnabled: bool, AutoApplyToNewTables: bool)")
	c.ShowMirroringTemplates = NewCommandSymbol("ShowMirroringTemplates", "(Name: string, Kind: string, ConnectionString: string, IsEnabled: bool, AutoApplyToNewTables: bool)")
	c.ApplyMirroringTemplate = NewCommandSymbol("ApplyMirroringTemplate", "(Name: string, Kind: string, ConnectionString: string, IsEnabled: bool, AutoApplyToNewTables: bool)")
	c.ShowTablePolicyPartitioning = NewCommandSymbol("ShowTablePolicyPartitioning", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTableStarPolicyPartitioning = NewCommandSymbol("ShowTableStarPolicyPartitioning", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablePolicyPartitioning = NewCommandSymbol("AlterTablePolicyPartitioning", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeTablePolicyPartitioning = NewCommandSymbol("AlterMergeTablePolicyPartitioning", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMaterializedViewPolicyPartitioning = NewCommandSymbol("AlterMaterializedViewPolicyPartitioning", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeMaterializedViewPolicyPartitioning = NewCommandSymbol("AlterMergeMaterializedViewPolicyPartitioning", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteTablePolicyPartitioning = NewCommandSymbol("DeleteTablePolicyPartitioning", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteMaterializedViewPolicyPartitioning = NewCommandSymbol("DeleteMaterializedViewPolicyPartitioning", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTablePolicyRestrictedViewAccess = NewCommandSymbol("ShowTablePolicyRestrictedViewAccess", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTableStarPolicyRestrictedViewAccess = NewCommandSymbol("ShowTableStarPolicyRestrictedViewAccess", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablePolicyRestrictedViewAccess = NewCommandSymbol("AlterTablePolicyRestrictedViewAccess", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablesPolicyRestrictedViewAccess = NewCommandSymbol("AlterTablesPolicyRestrictedViewAccess", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteTablePolicyRestrictedViewAccess = NewCommandSymbol("DeleteTablePolicyRestrictedViewAccess", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowClusterPolicyRowStore = NewCommandSymbol("ShowClusterPolicyRowStore", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterClusterPolicyRowStore = NewCommandSymbol("AlterClusterPolicyRowStore", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeClusterPolicyRowStore = NewCommandSymbol("AlterMergeClusterPolicyRowStore", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteClusterPolicyRowStore = NewCommandSymbol("DeleteClusterPolicyRowStore", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowClusterPolicySandbox = NewCommandSymbol("ShowClusterPolicySandbox", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterClusterPolicySandbox = NewCommandSymbol("AlterClusterPolicySandbox", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteClusterPolicySandbox = NewCommandSymbol("DeleteClusterPolicySandbox", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowClusterSandboxesStats = NewCommandSymbol("ShowClusterSandboxesStats", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowDatabasePolicySharding = NewCommandSymbol("ShowDatabasePolicySharding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTablePolicySharding = NewCommandSymbol("ShowTablePolicySharding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTableStarPolicySharding = NewCommandSymbol("ShowTableStarPolicySharding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterDatabasePolicySharding = NewCommandSymbol("AlterDatabasePolicySharding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablePolicySharding = NewCommandSymbol("AlterTablePolicySharding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeDatabasePolicySharding = NewCommandSymbol("AlterMergeDatabasePolicySharding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeTablePolicySharding = NewCommandSymbol("AlterMergeTablePolicySharding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteDatabasePolicySharding = NewCommandSymbol("DeleteDatabasePolicySharding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteTablePolicySharding = NewCommandSymbol("DeleteTablePolicySharding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterClusterPolicySharding = NewCommandSymbol("AlterClusterPolicySharding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeClusterPolicySharding = NewCommandSymbol("AlterMergeClusterPolicySharding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteClusterPolicySharding = NewCommandSymbol("DeleteClusterPolicySharding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowClusterPolicySharding = NewCommandSymbol("ShowClusterPolicySharding", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowDatabasePolicyShardsGrouping = NewCommandSymbol("ShowDatabasePolicyShardsGrouping", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterDatabasePolicyShardsGrouping = NewCommandSymbol("AlterDatabasePolicyShardsGrouping", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeDatabasePolicyShardsGrouping = NewCommandSymbol("AlterMergeDatabasePolicyShardsGrouping", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteDatabasePolicyShardsGrouping = NewCommandSymbol("DeleteDatabasePolicyShardsGrouping", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowDatabasePolicyStreamingIngestion = NewCommandSymbol("ShowDatabasePolicyStreamingIngestion", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTablePolicyStreamingIngestion = NewCommandSymbol("ShowTablePolicyStreamingIngestion", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowClusterPolicyStreamingIngestion = NewCommandSymbol("ShowClusterPolicyStreamingIngestion", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterDatabasePolicyStreamingIngestion = NewCommandSymbol("AlterDatabasePolicyStreamingIngestion", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeDatabasePolicyStreamingIngestion = NewCommandSymbol("AlterMergeDatabasePolicyStreamingIngestion", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablePolicyStreamingIngestion = NewCommandSymbol("AlterTablePolicyStreamingIngestion", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeTablePolicyStreamingIngestion = NewCommandSymbol("AlterMergeTablePolicyStreamingIngestion", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterClusterPolicyStreamingIngestion = NewCommandSymbol("AlterClusterPolicyStreamingIngestion", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeClusterPolicyStreamingIngestion = NewCommandSymbol("AlterMergeClusterPolicyStreamingIngestion", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteDatabasePolicyStreamingIngestion = NewCommandSymbol("DeleteDatabasePolicyStreamingIngestion", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteTablePolicyStreamingIngestion = NewCommandSymbol("DeleteTablePolicyStreamingIngestion", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteClusterPolicyStreamingIngestion = NewCommandSymbol("DeleteClusterPolicyStreamingIngestion", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowDatabasePolicyManagedIdentity = NewCommandSymbol("ShowDatabasePolicyManagedIdentity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowClusterPolicyManagedIdentity = NewCommandSymbol("ShowClusterPolicyManagedIdentity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterDatabasePolicyManagedIdentity = NewCommandSymbol("AlterDatabasePolicyManagedIdentity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeDatabasePolicyManagedIdentity = NewCommandSymbol("AlterMergeDatabasePolicyManagedIdentity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterClusterPolicyManagedIdentity = NewCommandSymbol("AlterClusterPolicyManagedIdentity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeClusterPolicyManagedIdentity = NewCommandSymbol("AlterMergeClusterPolicyManagedIdentity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteDatabasePolicyManagedIdentity = NewCommandSymbol("DeleteDatabasePolicyManagedIdentity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteClusterPolicyManagedIdentity = NewCommandSymbol("DeleteClusterPolicyManagedIdentity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTablePolicyAutoDelete = NewCommandSymbol("ShowTablePolicyAutoDelete", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablePolicyAutoDelete = NewCommandSymbol("AlterTablePolicyAutoDelete", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteTablePolicyAutoDelete = NewCommandSymbol("DeleteTablePolicyAutoDelete", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowClusterPolicyCallout = NewCommandSymbol("ShowClusterPolicyCallout", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterClusterPolicyCallout = NewCommandSymbol("AlterClusterPolicyCallout", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeClusterPolicyCallout = NewCommandSymbol("AlterMergeClusterPolicyCallout", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteClusterPolicyCallout = NewCommandSymbol("DeleteClusterPolicyCallout", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowClusterPolicyCapacity = NewCommandSymbol("ShowClusterPolicyCapacity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterClusterPolicyCapacity = NewCommandSymbol("AlterClusterPolicyCapacity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeClusterPolicyCapacity = NewCommandSymbol("AlterMergeClusterPolicyCapacity", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowClusterPolicyMaterializedViewPrioritization = NewCommandSymbol("ShowClusterPolicyMaterializedViewPrioritization", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string, EffectivePolicy: string)")
	c.AlterClusterPolicyMaterializedViewPrioritization = NewCommandSymbol("AlterClusterPolicyMaterializedViewPrioritization", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeClusterPolicyMaterializedViewPrioritization = NewCommandSymbol("AlterMergeClusterPolicyMaterializedViewPrioritization", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DropClusterPolicyMaterializedViewPrioritization = NewCommandSymbol("DropClusterPolicyMaterializedViewPrioritization", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowClusterPolicyRequestClassification = NewCommandSymbol("ShowClusterPolicyRequestClassification", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterClusterPolicyRequestClassification = NewCommandSymbol("AlterClusterPolicyRequestClassification", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeClusterPolicyRequestClassification = NewCommandSymbol("AlterMergeClusterPolicyRequestClassification", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteClusterPolicyRequestClassification = NewCommandSymbol("DeleteClusterPolicyRequestClassification", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowClusterPolicyMultiDatabaseAdmins = NewCommandSymbol("ShowClusterPolicyMultiDatabaseAdmins", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterClusterPolicyMultiDatabaseAdmins = NewCommandSymbol("AlterClusterPolicyMultiDatabaseAdmins", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeClusterPolicyMultiDatabaseAdmins = NewCommandSymbol("AlterMergeClusterPolicyMultiDatabaseAdmins", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowDatabasePolicyDiagnostics = NewCommandSymbol("ShowDatabasePolicyDiagnostics", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowClusterPolicyDiagnostics = NewCommandSymbol("ShowClusterPolicyDiagnostics", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterDatabasePolicyDiagnostics = NewCommandSymbol("AlterDatabasePolicyDiagnostics", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeDatabasePolicyDiagnostics = NewCommandSymbol("AlterMergeDatabasePolicyDiagnostics", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterClusterPolicyDiagnostics = NewCommandSymbol("AlterClusterPolicyDiagnostics", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeClusterPolicyDiagnostics = NewCommandSymbol("AlterMergeClusterPolicyDiagnostics", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteDatabasePolicyDiagnostics = NewCommandSymbol("DeleteDatabasePolicyDiagnostics", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowClusterPolicyQueryWeakConsistency = NewCommandSymbol("ShowClusterPolicyQueryWeakConsistency", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterClusterPolicyQueryWeakConsistency = NewCommandSymbol("AlterClusterPolicyQueryWeakConsistency", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterMergeClusterPolicyQueryWeakConsistency = NewCommandSymbol("AlterMergeClusterPolicyQueryWeakConsistency", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTablePolicyExtentTagsRetention = NewCommandSymbol("ShowTablePolicyExtentTagsRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowTableStarPolicyExtentTagsRetention = NewCommandSymbol("ShowTableStarPolicyExtentTagsRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowDatabasePolicyExtentTagsRetention = NewCommandSymbol("ShowDatabasePolicyExtentTagsRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterTablePolicyExtentTagsRetention = NewCommandSymbol("AlterTablePolicyExtentTagsRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.AlterDatabasePolicyExtentTagsRetention = NewCommandSymbol("AlterDatabasePolicyExtentTagsRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteTablePolicyExtentTagsRetention = NewCommandSymbol("DeleteTablePolicyExtentTagsRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.DeleteDatabasePolicyExtentTagsRetention = NewCommandSymbol("DeleteDatabasePolicyExtentTagsRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowPrincipalRoles = NewCommandSymbol("ShowPrincipalRoles", "(Scope: string, DisplayName: string, AADObjectID: string, Role: string)")
	c.ShowDatabasePrincipalRoles = NewCommandSymbol("ShowDatabasePrincipalRoles", "(Scope: string, DisplayName: string, AADObjectID: string, Role: string)")
	c.ShowTablePrincipalRoles = NewCommandSymbol("ShowTablePrincipalRoles", "(Scope: string, DisplayName: string, AADObjectID: string, Role: string)")
	c.ShowGraphModelPrincipalRoles = NewCommandSymbol("ShowGraphModelPrincipalRoles", "(Scope: string, DisplayName: string, AADObjectID: string, Role: string)")
	c.ShowExternalTablesPrincipalRoles = NewCommandSymbol("ShowExternalTablesPrincipalRoles", "(Scope: string, DisplayName: string, AADObjectID: string, Role: string)")
	c.ShowFunctionPrincipalRoles = NewCommandSymbol("ShowFunctionPrincipalRoles", "(Scope: string, DisplayName: string, AADObjectID: string, Role: string)")
	c.ShowClusterPrincipalRoles = NewCommandSymbol("ShowClusterPrincipalRoles", "(Scope: string, DisplayName: string, AADObjectID: string, Role: string)")
	c.ShowClusterPrincipals = NewCommandSymbol("ShowClusterPrincipals", "(Role: string, PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Notes: string, RoleAssignmentIdentifier: string)")
	c.ShowDatabasePrincipals = NewCommandSymbol("ShowDatabasePrincipals", "(Role: string, PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Notes: string, RoleAssignmentIdentifier: string)")
	c.ShowTablePrincipals = NewCommandSymbol("ShowTablePrincipals", "(Role: string, PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Notes: string, RoleAssignmentIdentifier: string)")
	c.ShowGraphModelPrincipals = NewCommandSymbol("ShowGraphModelPrincipals", "(Role: string, PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Notes: string, RoleAssignmentIdentifier: string)")
	c.ShowFunctionPrincipals = NewCommandSymbol("ShowFunctionPrincipals", "(Role: string, PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Notes: string, RoleAssignmentIdentifier: string)")
	c.AddClusterRole = NewCommandSymbol("AddClusterRole", "(Role: string, PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Notes: string, RoleAssignmentIdentifier: string)")
	c.DropClusterRole = NewCommandSymbol("DropClusterRole", "(Role: string, PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Notes: string, RoleAssignmentIdentifier: string)")
	c.SetClusterRole = NewCommandSymbol("SetClusterRole", "(Role: string, PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Notes: string, RoleAssignmentIdentifier: string)")
	c.AddDatabaseRole = NewCommandSymbol("AddDatabaseRole", "(Role: string, PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Notes: string, RoleAssignmentIdentifier: string)")
	c.DropDatabaseRole = NewCommandSymbol("DropDatabaseRole", "(Role: string, PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Notes: string, RoleAssignmentIdentifier: string)")
	c.SetDatabaseRole = NewCommandSymbol("SetDatabaseRole", "(Role: string, PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Notes: string, RoleAssignmentIdentifier: string)")
	c.AddTableRole = NewCommandSymbol("AddTableRole", "(Role: string, PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Notes: string, RoleAssignmentIdentifier: string)")
	c.DropTableRole = NewCommandSymbol("DropTableRole", "(Role: string, PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Notes: string, RoleAssignmentIdentifier: string)")
	c.SetTableRole = NewCommandSymbol("SetTableRole", "(Role: string, PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Notes: string, RoleAssignmentIdentifier: string)")
	c.AddFunctionRole = NewCommandSymbol("AddFunctionRole", "(Role: string, PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Notes: string, RoleAssignmentIdentifier: string)")
	c.DropFunctionRole = NewCommandSymbol("DropFunctionRole", "(Role: string, PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Notes: string, RoleAssignmentIdentifier: string)")
	c.SetFunctionRole = NewCommandSymbol("SetFunctionRole", "(Role: string, PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Notes: string, RoleAssignmentIdentifier: string)")
	c.ShowClusterBlockedPrincipals = NewCommandSymbol("ShowClusterBlockedPrincipals", "(PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Application: string, User: string, BlockedUntil: datetime, Reason: string)")
	c.AddClusterBlockedPrincipals = NewCommandSymbol("AddClusterBlockedPrincipals", "(PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Application: string, User: string, BlockedUntil: datetime, Reason: string)")
	c.DropClusterBlockedPrincipals = NewCommandSymbol("DropClusterBlockedPrincipals", "(PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Application: string, User: string, BlockedUntil: datetime, Reason: string)")
	c.SetClusterMaintenanceMode = NewCommandSymbol("SetClusterMaintenanceMode", "()")
	c.IngestIntoTable = NewCommandSymbol("IngestIntoTable", "(ExtentId: guid, ItemLoaded: string, Duration: string, HasErrors: string, OperationId: guid)")
	c.IngestInlineIntoTable = NewCommandSymbol("IngestInlineIntoTable", "(ExtentId: guid)")
	c.SetTable = NewCommandSymbol("SetTable", "(ExtentId: guid, OriginalSize: long, ExtentSize: long, ColumnSize: long, IndexSize: long, RowCount: long)")
	c.AppendTable = NewCommandSymbol("AppendTable", "(ExtentId: guid, OriginalSize: long, ExtentSize: long, ColumnSize: long, IndexSize: long, RowCount: long)")
	c.SetOrAppendTable = NewCommandSymbol("SetOrAppendTable", "(ExtentId: guid, OriginalSize: long, ExtentSize: long, ColumnSize: long, IndexSize: long, RowCount: long)")
	c.SetOrReplaceTable = NewCommandSymbol("SetOrReplaceTable", "(ExtentId: guid, OriginalSize: long, ExtentSize: long, ColumnSize: long, IndexSize: long, RowCount: long)")
	c.ExportToStorage = NewCommandSymbol("ExportToStorage", "()")
	c.ExportToSqlTable = NewCommandSymbol("ExportToSqlTable", "()")
	c.ExportToExternalTable = NewCommandSymbol("ExportToExternalTable", "()")
	c.CreateOrAlterContinuousExport = NewCommandSymbol("CreateOrAlterContinuousExport", "()")
	c.ShowContinuousExport = NewCommandSymbol("ShowContinuousExport", "(Name: string, ExternalTableName: string, Query: string, ForcedLatency: timespan, IntervalBetweenRuns: timespan, CursorScopedTables: string, ExportProperties: string, LastRunTime: datetime, StartCursor: string, IsDisabled: bool, LastRunResult: string, ExportedTo: datetime, IsRunning: bool)")
	c.ShowContinuousExports = NewCommandSymbol("ShowContinuousExports", "(Name: string, ExternalTableName: string, Query: string, ForcedLatency: timespan, IntervalBetweenRuns: timespan, CursorScopedTables: string, ExportProperties: string, LastRunTime: datetime, StartCursor: string, IsDisabled: bool, LastRunResult: string, ExportedTo: datetime, IsRunning: bool)")
	c.ShowClusterPendingContinuousExports = NewCommandSymbol("ShowClusterPendingContinuousExports", "(Name: string, ExternalTableName: string, Query: string, ForcedLatency: timespan, IntervalBetweenRuns: timespan, CursorScopedTables: string, ExportProperties: string, LastRunTime: datetime, StartCursor: string, IsDisabled: bool, LastRunResult: string, ExportedTo: datetime, IsRunning: bool)")
	c.ShowContinuousExportExportedArtifacts = NewCommandSymbol("ShowContinuousExportExportedArtifacts", "(Timestamp: datetime, ExternalTableName: string, Path: string, NumRecords: long)")
	c.ShowContinuousExportFailures = NewCommandSymbol("ShowContinuousExportFailures", "(Timestamp: datetime, OperationId: string, Name: string, LastSuccessRun: datetime, FailureKind: string, Details: string)")
	c.SetContinuousExportCursor = NewCommandSymbol("SetContinuousExportCursor", "()")
	c.DropContinuousExport = NewCommandSymbol("DropContinuousExport", "(Name: string, ExternalTableName: string, Query: string, ForcedLatency: timespan, IntervalBetweenRuns: timespan, CursorScopedTables: string, ExportProperties: string, LastRunTime: datetime, StartCursor: string, IsDisabled: bool, LastRunResult: string, ExportedTo: datetime, IsRunning: bool)")
	c.EnableContinuousExport = NewCommandSymbol("EnableContinuousExport", "(Name: string, ExternalTableName: string, Query: string, ForcedLatency: timespan, IntervalBetweenRuns: timespan, CursorScopedTables: string, ExportProperties: string, LastRunTime: datetime, StartCursor: string, IsDisabled: bool, LastRunResult: string, ExportedTo: datetime, IsRunning: bool)")
	c.DisableContinuousExport = NewCommandSymbol("DisableContinuousExport", "(Name: string, ExternalTableName: string, Query: string, ForcedLatency: timespan, IntervalBetweenRuns: timespan, CursorScopedTables: string, ExportProperties: string, LastRunTime: datetime, StartCursor: string, IsDisabled: bool, LastRunResult: string, ExportedTo: datetime, IsRunning: bool)")
	c.CreateMaterializedView = NewCommandSymbol("CreateMaterializedView", "()")
	c.CreateMaterializedViewOverMaterializedView = NewCommandSymbol("CreateMaterializedViewOverMaterializedView", "()")
	c.RenameMaterializedView = NewCommandSymbol("RenameMaterializedView", "(Name: string, SourceTable: string, Query: string, MaterializedTo: datetime, LastRun: datetime, LastRunResult: string, IsHealthy: bool, IsEnabled: bool, Folder: string, DocString: string, AutoUpdateSchema: bool, EffectiveDateTime: datetime, Lookback:timespan, LookbackColumn:string)")
	c.ShowMaterializedView = NewCommandSymbol("ShowMaterializedView", "(Name: string, SourceTable: string, Query: string, MaterializedTo: datetime, LastRun: datetime, LastRunResult: string, IsHealthy: bool, IsEnabled: bool, Folder: string, DocString: string, AutoUpdateSchema: bool, EffectiveDateTime: datetime, Lookback:timespan, LookbackColumn:string)")
	c.ShowMaterializedViews = NewCommandSymbol("ShowMaterializedViews", "(Name: string, SourceTable: string, Query: string, MaterializedTo: datetime, LastRun: datetime, LastRunResult: string, IsHealthy: bool, IsEnabled: bool, Folder: string, DocString: string, AutoUpdateSchema: bool, EffectiveDateTime: datetime, Lookback:timespan, LookbackColumn:string)")
	c.ShowMaterializedViewsDetails = NewCommandSymbol("ShowMaterializedViewsDetails", "(MaterializedViewName: string, DatabaseName: string, Folder: string, DocString: string, TotalExtents: long, TotalExtentSize: real, TotalOriginalSize: real, TotalRowCount: long, HotExtents: long, HotExtentSize: real, HotOriginalSize: real, HotRowCount: long, AuthorizedPrincipals: string, RetentionPolicy: string, CachingPolicy: string, ShardingPolicy: string, MergePolicy: string, MinExtentsCreationTime: datetime, MaxExtentsCreationTime: datetime)")
	c.ShowMaterializedViewDetails = NewCommandSymbol("ShowMaterializedViewDetails", "(MaterializedViewName: string, DatabaseName: string, Folder: string, DocString: string, TotalExtents: long, TotalExtentSize: real, TotalOriginalSize: real, TotalRowCount: long, HotExtents: long, HotExtentSize: real, HotOriginalSize: real, HotRowCount: long, AuthorizedPrincipals: string, RetentionPolicy: string, CachingPolicy: string, ShardingPolicy: string, MergePolicy: string, MinExtentsCreationTime: datetime, MaxExtentsCreationTime: datetime)")
	c.ShowMaterializedViewPolicyRetention = NewCommandSymbol("ShowMaterializedViewPolicyRetention", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowMaterializedViewPolicyMerge = NewCommandSymbol("ShowMaterializedViewPolicyMerge", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowMaterializedViewPolicyPartitioning = NewCommandSymbol("ShowMaterializedViewPolicyPartitioning", "(PolicyName: string, EntityName: string, Policy: string, ChildEntities: string, EntityType: string)")
	c.ShowMaterializedViewExtents = NewCommandSymbol("ShowMaterializedViewExtents", "(ExtentId: guid, DatabaseName: string, TableName: string, MaxCreatedOn: datetime, OriginalSize: real, ExtentSize: real, CompressedSize: real, IndexSize: real, Blocks: long, Segments: long, ExtentContainerId: string, RowCount: long, MinCreatedOn: datetime, Tags: string, Kind: string, DeletedRowCount: long)")
	c.AlterMaterializedView = NewCommandSymbol("AlterMaterializedView", "(Name: string, SourceTable: string, Query: string, MaterializedTo: datetime, LastRun: datetime, LastRunResult: string, IsHealthy: bool, IsEnabled: bool, Folder: string, DocString: string, AutoUpdateSchema: bool, EffectiveDateTime: datetime, Lookback:timespan, LookbackColumn:string)")
	c.AlterMaterializedViewOverMaterializedView = NewCommandSymbol("AlterMaterializedViewOverMaterializedView", "(Name: string, SourceTable: string, Query: string, MaterializedTo: datetime, LastRun: datetime, LastRunResult: string, IsHealthy: bool, IsEnabled: bool, Folder: string, DocString: string, AutoUpdateSchema: bool, EffectiveDateTime: datetime, Lookback:timespan, LookbackColumn:string)")
	c.CreateOrAlterMaterializedView = NewCommandSymbol("CreateOrAlterMaterializedView", "(Name: string, SourceTable: string, Query: string, MaterializedTo: datetime, LastRun: datetime, LastRunResult: string, IsHealthy: bool, IsEnabled: bool, Folder: string, DocString: string, AutoUpdateSchema: bool, EffectiveDateTime: datetime, Lookback:timespan, LookbackColumn:string)")
	c.CreateOrAlterMaterializedViewOverMaterializedView = NewCommandSymbol("CreateOrAlterMaterializedViewOverMaterializedView", "(Name: string, SourceTable: string, Query: string, MaterializedTo: datetime, LastRun: datetime, LastRunResult: string, IsHealthy: bool, IsEnabled: bool, Folder: string, DocString: string, AutoUpdateSchema: bool, EffectiveDateTime: datetime, Lookback:timespan, LookbackColumn:string)")
	c.DropMaterializedView = NewCommandSymbol("DropMaterializedView", "(Name: string, SourceTable: string, Query: string, MaterializedTo: datetime, LastRun: datetime, LastRunResult: string, IsHealthy: bool, IsEnabled: bool, Folder: string, DocString: string, AutoUpdateSchema: bool, EffectiveDateTime: datetime, Lookback:timespan, LookbackColumn:string)")
	c.EnableDisableMaterializedView = NewCommandSymbol("EnableDisableMaterializedView", "(Name: string, SourceTable: string, Query: string, MaterializedTo: datetime, LastRun: datetime, LastRunResult: string, IsHealthy: bool, IsEnabled: bool, Folder: string, DocString: string, AutoUpdateSchema: bool, EffectiveDateTime: datetime, Lookback:timespan, LookbackColumn:string)")
	c.ShowMaterializedViewPrincipals = NewCommandSymbol("ShowMaterializedViewPrincipals", "(Role: string, PrincipalType: string, PrincipalDisplayName: string, PrincipalObjectId: string, PrincipalFQN: string, Notes: string, RoleAssignmentIdentifier: string)")
	c.ShowMaterializedViewSchemaAsJson = NewCommandSymbol("ShowMaterializedViewSchemaAsJson", "(TableName: string, Schema: string, DatabaseName: string, Folder: string, DocString: string)")
	c.ShowMaterializedViewCslSchema = NewCommandSymbol("ShowMaterializedViewCslSchema", "(TableName: string, Schema: string, DatabaseName: string, Folder: string, DocString: string)")
	c.AlterMaterializedViewFolder = NewCommandSymbol("AlterMaterializedViewFolder", "(Name: string, SourceTable: string, Query: string, MaterializedTo: datetime, LastRun: datetime, LastRunResult: string, IsHealthy: bool, IsEnabled: bool, Folder: string, DocString: string, AutoUpdateSchema: bool, EffectiveDateTime: datetime, Lookback:timespan, LookbackColumn:string)")
	c.AlterMaterializedViewDocString = NewCommandSymbol("AlterMaterializedViewDocString", "(Name: string, SourceTable: string, Query: string, MaterializedTo: datetime, LastRun: datetime, LastRunResult: string, IsHealthy: bool, IsEnabled: bool, Folder: string, DocString: string, AutoUpdateSchema: bool, EffectiveDateTime: datetime, Lookback:timespan, LookbackColumn:string)")
	c.AlterMaterializedViewLookback = NewCommandSymbol("AlterMaterializedViewLookback", "(Name: string, SourceTable: string, Query: string, MaterializedTo: datetime, LastRun: datetime, LastRunResult: string, IsHealthy: bool, IsEnabled: bool, Folder: string, DocString: string, AutoUpdateSchema: bool, EffectiveDateTime: datetime, Lookback:timespan, LookbackColumn:string)")
	c.AlterMaterializedViewAutoUpdateSchema = NewCommandSymbol("AlterMaterializedViewAutoUpdateSchema", "(Name: string, SourceTable: string, Query: string, MaterializedTo: datetime, LastRun: datetime, LastRunResult: string, IsHealthy: bool, IsEnabled: bool, Folder: string, DocString: string, AutoUpdateSchema: bool, EffectiveDateTime: datetime, Lookback:timespan, LookbackColumn:string)")
	c.ClearMaterializedViewData = NewCommandSymbol("ClearMaterializedViewData", "(ExtentId: guid, TableName: string, CreatedOn: datetime)")
	c.SetMaterializedViewCursor = NewCommandSymbol("SetMaterializedViewCursor", "()")
	c.ShowTableOperationsMirroringStatus = NewCommandSymbol("ShowTableOperationsMirroringStatus", "(TableName: string, IsEnabled: bool, ExportProperties: string, ManagedIdentityIdentifier: string, IsExportRunning: bool, LastExportStartTime: datetime, LastExportResult: string, LastExportedDataTime: datetime)")
	c.ShowTableOperationsMirroringExportedArtifacts = NewCommandSymbol("ShowTableOperationsMirroringExportedArtifacts", "(Timestamp: datetime, ExternalTableName: string, Path: string, NumRecords: long)")
	c.ShowTableOperationsMirroringFailures = NewCommandSymbol("ShowTableOperationsMirroringFailures", "(Timestamp: datetime, OperationId: string, Name: string, LastSuccessRun: datetime, FailureKind: string, Details: string)")
	c.ShowCluster = NewCommandSymbol("ShowCluster", "(NodeId: string, Address: string, Name: string, StartTime: datetime, IsAdmin: bool, MachineTotalMemory: long, MachineAvailableMemory: long, ProcessorCount: int, EnvironmentDescription: string)")
	c.ShowClusterDetails = NewCommandSymbol("ShowClusterDetails", "(NodeId: string, Address: string, Name: string, StartTime: datetime, IsAdmin: bool, MachineTotalMemory: long, MachineAvailableMemory: long, ProcessorCount: int, EnvironmentDescription: string, NetAndClsVersion: string)")
	c.ShowDiagnostics = NewCommandSymbol("ShowDiagnostics", "(IsHealthy: bool, EnvironmentDescription: string, IsScaleOutRequired: bool, MachinesTotal: int, MachinesOffline: int, NodeLastRestartedOn: datetime, AdminLastElectedOn: datetime, ExtentsTotal: int, DiskColdAllocationPercentage: int, InstancesTargetBasedOnDataCapacity: int, TotalOriginalDataSize: real, TotalExtentSize: real, IngestionsLoadFactor: real, IngestionsInProgress: long, IngestionsSuccessRate: real, MergesInProgress: long, BuildVersion: string, BuildTime: datetime, ClusterDataCapacityFactor: real, IsDataWarmingRequired: bool, DataWarmingLastRunOn: datetime, MergesSuccessRate: real, NotHealthyReason: string, IsAttentionRequired: bool, AttentionRequiredReason: string, ProductVersion: string, FailedIngestOperations: int, FailedMergeOperations: int, MaxExtentsInSingleTable: int, TableWithMaxExtents: string, MergesLoadFactor: double, WarmExtentSize: real, NumberOfDatabases: int, PurgeExtentsRebuildLoadFactor: real, PurgeExtentsRebuildInProgress: long, PurgesInProgress: long, RowStoreLocalStorageCapacityFactor: real, ExportsLoadFactor: real, ExportsInProgress: long, PendingContinuousExports: long, MaxContinuousExportLatenessMinutes: long, RowStoreSealsInProgress: long, IsRowStoreUnhealthy: bool, MachinesSuspended: int, DataPartitioningLoadFactor: real, DataPartitioningOperationsInProgress: long, MinPartitioningPercentageInSingleTable: real, TableWithMinPartitioningPercentage: string, V3DataCapacityFactor: real, MaterializedViewsInProgress: long, IngestionCapacityUtilization: real, ShardsWarmingStatus: string, ShardsWarmingTemperature: real, ShardsWarmingDetails: string, StoredQueryResultsInProgress: long, HotDataDiskSpaceUsage: real, MirroringOperationsLoadFactor: real, MirroringOperationsInProgress: long, QueryAccelerationLoadFactor: real, QueryAccelerationInProgress: long, QueryAccelerationCapacityUtilization: real, GraphSnapshotsLoadFactor: real, GraphSnapshotsInProgress: long")
	c.ShowCapacity = NewCommandSymbol("ShowCapacity", "(Resource: string, Total: long, Consumed: long, Remaining: long)")
	c.ShowOperations = NewCommandSymbol("ShowOperations", "(OperationId: guid, Operation: string, NodeId: string, StartedOn: datetime, LastUpdatedOn: datetime, Duration: timespan, State: string, Status: string, RootActivityId: guid, ShouldRetry: bool, Database: string, Principal: string, User: string, AdminEpochStartTime: datetime, VirtualCluster: string)")
	c.ShowOperationDetails = NewCommandSymbol("ShowOperationDetails", "()")
	c.ShowJournal = NewCommandSymbol("ShowJournal", "(Event: string, EventTimestamp: datetime, Database: string, EntityName: string, UpdatedEntityName: string, EntityVersion: string, EntityContainerName: string, OriginalEntityState: string, UpdatedEntityState: string, ChangeCommand: string, Principal: string, RootActivityId: guid, ClientRequestId: string, User: string, OriginalEntityVersion: string)")
	c.ShowDatabaseJournal = NewCommandSymbol("ShowDatabaseJournal", "(Event: string, EventTimestamp: datetime, Database: string, EntityName: string, UpdatedEntityName: string, EntityVersion: string, EntityContainerName: string, OriginalEntityState: string, UpdatedEntityState: string, ChangeCommand: string, Principal: string, RootActivityId: guid, ClientRequestId: string, User: string, OriginalEntityVersion: string)")
	c.ShowClusterJournal = NewCommandSymbol("ShowClusterJournal", "(Event: string, EventTimestamp: datetime, Database: string, EntityName: string, UpdatedEntityName: string, EntityVersion: string, EntityContainerName: string, OriginalEntityState: string, UpdatedEntityState: string, ChangeCommand: string, Principal: string, RootActivityId: guid, ClientRequestId: string, User: string, OriginalEntityVersion: string)")
	c.ShowQueries = NewCommandSymbol("ShowQueries", "(ClientActivityId: string, Text: string, Database: string, StartedOn: datetime, LastUpdatedOn: datetime, Duration: timespan, State: string, RootActivityId: guid, User: string, FailureReason: string, TotalCpu: timespan, CacheStatistics: dynamic, Application: string, MemoryPeak: long, ScannedExtentsStatistics: dynamic, Principal: string, ClientRequestProperties: dynamic, ResultSetStatistics: dynamic, WorkloadGroup: string, OverallQueryStats: string, VirtualCluster: string)")
	c.ShowRunningQueries = NewCommandSymbol("ShowRunningQueries", "(ClientActivityId: string, Text: string, Database: string, StartedOn: datetime, LastUpdatedOn: datetime, Duration: timespan, State: string, RootActivityId: guid, User: string, FailureReason: string, TotalCpu: timespan, CacheStatistics: dynamic, Application: string, MemoryPeak: long, ScannedEventStatistics: dynamic, Pricipal: string, ClientRequestProperties: dynamic, ResultSetStatistics: dynamic, WorkloadGroup: string)")
	c.CancelQuery = NewCommandSymbol("CancelQuery", "()")
	c.ShowQueryPlan = NewCommandSymbol("ShowQueryPlan", "(ResultType: string, Format: string, Content: string)")
	c.ShowCache = NewCommandSymbol("ShowCache", "(NodeId: string, TotalMemoryCapacity: long, MemoryCacheCapacity: long, MemoryCacheInUse: long, MemoryCacheHitCount: long, TotalDiskCapacity: long, DiskCacheCapacity: long, DiskCacheInUse: long, DiskCacheHitCount: long, DiskCacheMissCount: long, MemoryCacheDetails: string, DiskCacheDetails: string, WarmedShardsSize: long, ShardsSizeToWarm: long, HotShardsSize: long)")
	c.AlterCache = NewCommandSymbol("AlterCache", "()")
	c.ShowCommands = NewCommandSymbol("ShowCommands", "(ClientActivityId: string, CommandType: string, Text: string, Database: string, StartedOn: datetime, LastUpdatedOn: datetime, Duration: timespan, State: string, RootActivityId: guid, User: string, FailureReason: string, Application: string, Principal: string, TotalCpu: timespan, ResourcesUtilization: dynamic, ClientRequestProperties: dynamic, WorkloadGroup: string, VirtualCluster: string)")
	c.ShowCommandsAndQueries = NewCommandSymbol("ShowCommandsAndQueries", "(ClientActivityId: string, CommandType: string, Text: string, Database: string, StartedOn: datetime, LastUpdatedOn: datetime, Duration: timespan, State: string, FailureReason: string, RootActivityId: guid, User: string, Application: string, Principal: string, ClientRequestProperties: dynamic, TotalCpu: timespan, MemoryPeak: long, CacheStatistics: dynamic, ScannedExtentsStatistics: dynamic, ResultSetStatistics: dynamic, WorkloadGroup: string, VirtualCluster: string)")
	c.ShowIngestionFailures = NewCommandSymbol("ShowIngestionFailures", "(OperationId: guid, Database: string, Table: string, FailedOn: datetime, IngestionSourcePath: string, Details: string, FailureKind: string, RootActivityId: guid, OperationKind: string, OriginatesFromUpdatePolicy: bool, ErrorCode: string, Principal: string, ShouldRetry: bool, User: string, IngestionProperties: string)")
	c.ShowDataOperations = NewCommandSymbol("ShowDataOperations", "(Timestamp: datetime, Database: string, Table: string, ClientActivityId: string, OperationKind: string, OriginalSize: long, ExtentSize: long, RowCount: long, Duration: timespan, TotalCpu: timespan, ExtentCount: long, Principal: string, Properties: string)")
	c.ShowDatabaseKeyVaultSecrets = NewCommandSymbol("ShowDatabaseKeyVaultSecrets", "(KeyVaultSecretId: string)")
	c.ShowClusterExtents = NewCommandSymbol("ShowClusterExtents", "(ExtentId: guid, DatabaseName: string, TableName: string, MaxCreatedOn: datetime, OriginalSize: real, ExtentSize: real, CompressedSize: real, IndexSize: real, Blocks: long, Segments: long, ExtentContainerId: string, RowCount: long, MinCreatedOn: datetime, Tags: string, Kind: string, DeletedRowCount: long)")
	c.ShowClusterExtentsMetadata = NewCommandSymbol("ShowClusterExtentsMetadata", "(ExtentId: guid, DatabaseName: string, TableName: string, ExtentMetadata: string)")
	c.ShowDatabaseExtents = NewCommandSymbol("ShowDatabaseExtents", "(ExtentId: guid, DatabaseName: string, TableName: string, MaxCreatedOn: datetime, OriginalSize: real, ExtentSize: real, CompressedSize: real, IndexSize: real, Blocks: long, Segments: long, ExtentContainerId: string, RowCount: long, MinCreatedOn: datetime, Tags: string, Kind: string, DeletedRowCount: long)")
	c.ShowDatabaseExtentsMetadata = NewCommandSymbol("ShowDatabaseExtentsMetadata", "(ExtentId: guid, DatabaseName: string, TableName: string, ExtentMetadata: string)")
	c.ShowDatabaseExtentTagsStatistics = NewCommandSymbol("ShowDatabaseExtentTagsStatistics", "(TableName: string, TotalExtentsCount: long, TaggedExtentsCount: long, TotalTagsCount: long, TotalTagsLength: long, DropByTagsCount: long, DropByTagsLength: long, IngestByTagsCount: long, IngestByTagsLength: long)")
	c.ShowDatabaseExtentsPartitioningStatistics = NewCommandSymbol("ShowDatabaseExtentsPartitioningStatistics", "(TableName: string, PartitioningPolicy: dynamic, TotalRowCount: long, PartitionedRowCount: long, PartitionedRowsPercentage: double)")
	c.ShowTableExtents = NewCommandSymbol("ShowTableExtents", "(ExtentId: guid, DatabaseName: string, TableName: string, MaxCreatedOn: datetime, OriginalSize: real, ExtentSize: real, CompressedSize: real, IndexSize: real, Blocks: long, Segments: long, ExtentContainerId: string, RowCount: long, MinCreatedOn: datetime, Tags: string, Kind: string, DeletedRowCount: long)")
	c.ShowTableExtentsMetadata = NewCommandSymbol("ShowTableExtentsMetadata", "(ExtentId: guid, DatabaseName: string, TableName: string, ExtentMetadata: string)")
	c.TableShardsGroupShow = NewCommandSymbol("TableShardsGroupShow", "(ExtentId: guid, DatabaseName: string, TableName: string, MaxCreatedOn: datetime, OriginalSize: real, ExtentSize: real, CompressedSize: real, IndexSize: real, Blocks: long, Segments: long, ExtentContainerId: string, RowCount: long, MinCreatedOn: datetime, Tags: string, Kind: string, DeletedRowCount: long)")
	c.TableShardsGroupMetadataShow = NewCommandSymbol("TableShardsGroupMetadataShow", "(ExtentId: guid, DatabaseName: string, TableName: string, ExtentMetadata: string)")
	c.TableShardGroupsShow = NewCommandSymbol("TableShardGroupsShow", "(Id: guid, ShardCount: long, RowCount: long, OriginalSize: long, ShardSize: long, CompressedSize: long, IndexSize: long, DeletedRowCount: long, PartitionedRowCount: long, DateTimeColumnRanges: dynamic, Partition: dynamic, CreationTimeRange: dynamic)")
	c.TableShardGroupsStatisticsShow = NewCommandSymbol("TableShardGroupsStatisticsShow", "(TableId: guid, long ShardGroupCount: long, ShardCount: long, RowCount: long, OriginalSize: long, ShardSize: long, CompressedSize: long, IndexSize: long, DeletedRowCount: long, PartitionedRowCount: long)")
	c.TablesShardGroupsStatisticsShow = NewCommandSymbol("TablesShardGroupsStatisticsShow", "(TableId: guid, long ShardGroupCount: long, ShardCount: long, RowCount: long, OriginalSize: long, ShardSize: long, CompressedSize: long, IndexSize: long, DeletedRowCount: long, PartitionedRowCount: long)")
	c.DatabaseShardGroupsStatisticsShow = NewCommandSymbol("DatabaseShardGroupsStatisticsShow", "(TableId: guid, long ShardGroupCount: long, ShardCount: long, RowCount: long, OriginalSize: long, ShardSize: long, CompressedSize: long, IndexSize: long, DeletedRowCount: long, PartitionedRowCount: long)")
	c.MergeExtents = NewCommandSymbol("MergeExtents", "(OriginalExtentId: string, ResultExtentId: string, Duration: timespan)")
	c.MergeExtentsDryrun = NewCommandSymbol("MergeExtentsDryrun", "(OriginalExtentId: string, ResultExtentId: string, Duration: timespan)")
	c.MoveExtentsFrom = NewCommandSymbol("MoveExtentsFrom", "(OriginalExtentId: string, ResultExtentId: string, Details: string)")
	c.MoveExtentsQuery = NewCommandSymbol("MoveExtentsQuery", "(OriginalExtentId: string, ResultExtentId: string, Details: string)")
	c.TableShuffleExtents = NewCommandSymbol("TableShuffleExtents", "(OriginalExtentId: string, ResultExtentId: string, Details: string)")
	c.TableShuffleExtentsQuery = NewCommandSymbol("TableShuffleExtentsQuery", "(OriginalExtentId: string, ResultExtentId: string, Details: string)")
	c.ReplaceExtents = NewCommandSymbol("ReplaceExtents", "(OriginalExtentId: string, ResultExtentId: string, Details: string)")
	c.DropExtent = NewCommandSymbol("DropExtent", "(ExtentId: guid, TableName: string, CreatedOn: datetime)")
	c.DropExtents = NewCommandSymbol("DropExtents", "(ExtentId: guid, TableName: string, CreatedOn: datetime)")
	c.DropExtentsPartitionMetadata = NewCommandSymbol("DropExtentsPartitionMetadata", "()")
	c.DropPretendExtentsByProperties = NewCommandSymbol("DropPretendExtentsByProperties", "(ExtentId: guid, TableName: string, CreatedOn: datetime)")
	c.ShowVersion = NewCommandSymbol("ShowVersion", "(BuildVersion: string, BuildTime: datetime, ServiceType: string, ProductVersion: string)")
	c.ClearTableData = NewCommandSymbol("ClearTableData", "(Status: string)")
	c.ClearTableCacheStreamingIngestionSchema = NewCommandSymbol("ClearTableCacheStreamingIngestionSchema", "(NodeId: string, Status: string)")
	c.ShowStorageArtifactsCleanupState = NewCommandSymbol("ShowStorageArtifactsCleanupState", "(DatabaseName: string, StorageAccount: string, ContainerId: guid, ContainerName: string, CleanupStateBlobName: string, LastModified: datetime, SizeInBytes: long, Content: string)")
	c.ClusterDropStorageArtifactsCleanupState = NewCommandSymbol("ClusterDropStorageArtifactsCleanupState", "(ContainerId: guid, ContainerName: string, FileName: string, Succeeded: bool, FailureDetails: string)")
	c.StoredQueryResultSet = NewCommandSymbol("StoredQueryResultSet", "()")
	c.StoredQueryResultSetOrReplace = NewCommandSymbol("StoredQueryResultSetOrReplace", "()")
	c.StoredQueryResultsShow = NewCommandSymbol("StoredQueryResultsShow", "(StoredQueryResultId:guid, Name:string, DatabaseName:string, PrincipalIdentity:string, SizeInBytes:long, RowCount:long, CreatedOn:datetime, ExpiresOn:datetime)")
	c.StoredQueryResultShowSchema = NewCommandSymbol("StoredQueryResultShowSchema", "(StoredQueryResult:string, Schema:string)")
	c.StoredQueryResultDrop = NewCommandSymbol("StoredQueryResultDrop", "(StoredQueryResultId:guid, Name:string, DatabaseName:string, PrincipalIdentity:string, SizeInBytes:long, RowCount:long, CreatedOn:datetime, ExpiresOn:datetime)")
	c.StoredQueryResultsDrop = NewCommandSymbol("StoredQueryResultsDrop", "(StoredQueryResultId:guid, Name:string, DatabaseName:string, PrincipalIdentity:string, SizeInBytes:long, RowCount:long, CreatedOn:datetime, ExpiresOn:datetime)")
	c.GraphModelCreateOrAlter = NewCommandSymbol("GraphModelCreateOrAlter", "(Name:string, CreationTime:datetime, Id:guid, SnapshotCount:long, Model:string, RetentionPolicy:string, CachingPolicy:string)")
	c.GraphModelShow = NewCommandSymbol("GraphModelShow", "()")
	c.GraphModelsShow = NewCommandSymbol("GraphModelsShow", "()")
	c.GraphModelDrop = NewCommandSymbol("GraphModelDrop", "(Name:string, CreationTime:datetime, Id:guid)")
	c.SetGraphModelAdmins = NewCommandSymbol("SetGraphModelAdmins", "()")
	c.AddGraphModelAdmins = NewCommandSymbol("AddGraphModelAdmins", "()")
	c.DropGraphModelAdmins = NewCommandSymbol("DropGraphModelAdmins", "()")
	c.GraphSnapshotMake = NewCommandSymbol("GraphSnapshotMake", "(Name:string, SnapshotTime:datetime, ModelName:string, ModelId:guid, ModelCreationTime:datetime, NodesCount: long, EdgesCount: long, RetentionPolicy:string, CachingPolicy:string)")
	c.GraphSnapshotShow = NewCommandSymbol("GraphSnapshotShow", "()")
	c.GraphSnapshotsShow = NewCommandSymbol("GraphSnapshotsShow", "()")
	c.GraphSnapshotDrop = NewCommandSymbol("GraphSnapshotDrop", "(Name:string, SnapshotTime:datetime, ModelName:string, ModelId:guid, ModelCreationTime:datetime)")
	c.GraphSnapshotsDrop = NewCommandSymbol("GraphSnapshotsDrop", "(Name:string, SnapshotTime:datetime, ModelName:string, ModelId:guid, ModelCreationTime:datetime)")
	c.GraphSnapshotShowDataStatistics = NewCommandSymbol("GraphSnapshotShowDataStatistics", "(DatabaseName:string, GraphModelName:string, SnapshotName:string, EntityKind:string, TotalExtentSize:long, OriginalSize:long, DataCompressedSize:long, IndexSize:long, CompressionRatio:real, TotalElementCount:long)")
	c.GraphSnapshotsShowDataStatistics = NewCommandSymbol("GraphSnapshotsShowDataStatistics", "(DatabaseName:string, GraphModelName:string, SnapshotName:string, EntityKind:string, TotalExtentSize:long, OriginalSize:long, DataCompressedSize:long, IndexSize:long, CompressionRatio:real, TotalElementCount:long)")
	c.GraphSnapshotShowStatistics = NewCommandSymbol("GraphSnapshotShowStatistics", "(DatabaseName:string, Name:string, SnapshotTime:datetime, ModelName:string, ModelId:guid, TotalCpu:timespan, MemoryPeak:long, Duration:timespan, Details:string, NodesCount:long, EdgesCount:long, NodesSize:long, EdgesSize:long)")
	c.GraphSnapshotsShowStatistics = NewCommandSymbol("GraphSnapshotsShowStatistics", "(DatabaseName:string, Name:string, SnapshotTime:datetime, ModelName:string, ModelId:guid, TotalCpu:timespan, MemoryPeak:long, Duration:timespan, Details:string, NodesCount:long, EdgesCount:long, NodesSize:long, EdgesSize:long)")
	c.GraphSnapshotShowFailures = NewCommandSymbol("GraphSnapshotShowFailures", "(OperationId:guid, DatabaseName:string, Name:string, SnapshotTime:datetime, ModelName:string, ModelId:guid, TotalCpu:timespan, MemoryPeak:long, Duration:timespan, Details:string, FailureReason:string, FailureKind:string)")
	c.GraphShardMake = NewCommandSymbol("GraphShardMake", "(Name:string, SnapshotTime:datetime, ModelName:string, ModelId:guid, ModelCreationTime:datetime, NodesCount: long, EdgesCount: long, RetentionPolicy:string, CachingPolicy:string)")
	c.GraphShardDrop = NewCommandSymbol("GraphShardDrop", "(Name:string, SnapshotTime:datetime, ModelName:string, ModelId:guid, ModelCreationTime:datetime, NodesCount: long, EdgesCount: long, RetentionPolicy:string, CachingPolicy:string)")
	c.ShowCertificates = NewCommandSymbol("ShowCertificates", "()")
	c.ShowClockDiagnostics = NewCommandSymbol("ShowClockDiagnostics", "()")
	c.ShowCloudSettings = NewCommandSymbol("ShowCloudSettings", "()")
	c.ShowCommConcurrency = NewCommandSymbol("ShowCommConcurrency", "()")
	c.ShowCommPools = NewCommandSymbol("ShowCommPools", "()")
	c.ShowFabricCache = NewCommandSymbol("ShowFabricCache", "()")
	c.ShowFabricLocks = NewCommandSymbol("ShowFabricLocks", "()")
	c.ShowFabricClocks = NewCommandSymbol("ShowFabricClocks", "()")
	c.ShowFeatureFlags = NewCommandSymbol("ShowFeatureFlags", "()")
	c.ShowMemPools = NewCommandSymbol("ShowMemPools", "()")
	c.ShowServicePoints = NewCommandSymbol("ShowServicePoints", "()")
	c.ShowTcpConnections = NewCommandSymbol("ShowTcpConnections", "()")
	c.ShowTcpPorts = NewCommandSymbol("ShowTcpPorts", "()")
	c.ShowThreadPools = NewCommandSymbol("ShowThreadPools", "()")
	c.ShowWhatsGoingOn = NewCommandSymbol("ShowWhatsGoingOn", "()")
	c.ExecuteDatabaseScript = NewCommandSymbol("ExecuteDatabaseScript", "()")
	c.ExecuteClusterScript = NewCommandSymbol("ExecuteClusterScript", "()")
	c.CreateRequestMemoryProfile = NewCommandSymbol("CreateRequestMemoryProfile", "()")
	c.CreateRequestSupport = NewCommandSymbol("CreateRequestSupport", "()")
	c.ShowRequestSupport = NewCommandSymbol("ShowRequestSupport", "()")
	c.ShowClusterAdminState = NewCommandSymbol("ShowClusterAdminState", "()")
	c.ClearRemoteClusterDatabaseSchema = NewCommandSymbol("ClearRemoteClusterDatabaseSchema", "()")
	c.ShowClusterMonitoring = NewCommandSymbol("ShowClusterMonitoring", "()")
	c.ShowClusterScaleIn = NewCommandSymbol("ShowClusterScaleIn", "()")
	c.ShowClusterServices = NewCommandSymbol("ShowClusterServices", "(NodeId: string, IsClusterAdmin: bool, IsFabricManager: bool, IsDatabaseAdmin: bool, IsWeakConsistencyNode: bool, IsRowStoreHostNode: bool, IsDataNode: bool)")
	c.ShowClusterNetwork = NewCommandSymbol("ShowClusterNetwork", "()")
	c.AlterClusterStorageKeys = NewCommandSymbol("AlterClusterStorageKeys", "()")
	c.ShowClusterStorageKeysHash = NewCommandSymbol("ShowClusterStorageKeysHash", "()")
	c.AlterFabricServiceAssignmentsCommand = NewCommandSymbol("AlterFabricServiceAssignmentsCommand", "()")
	c.DropFabricServiceAssignmentsCommand = NewCommandSymbol("DropFabricServiceAssignmentsCommand", "()")
	c.CreateEntityGroupCommand = NewCommandSymbol("CreateEntityGroupCommand", "(Name: string, Entities: string)")
	c.CreateOrAlterEntityGroupCommand = NewCommandSymbol("CreateOrAlterEntityGroupCommand", "(Name: string, Entities: string)")
	c.AlterEntityGroup = NewCommandSymbol("AlterEntityGroup", "(Name: string, Entities: string)")
	c.AlterMergeEntityGroup = NewCommandSymbol("AlterMergeEntityGroup", "(Name: string, Entities: string)")
	c.DropEntityGroup = NewCommandSymbol("DropEntityGroup", "(Name: string, Entities: string)")
	c.ShowEntityGroup = NewCommandSymbol("ShowEntityGroup", "(Name: string, Entities: string)")
	c.ShowEntityGroups = NewCommandSymbol("ShowEntityGroups", "(Name: string, Entities: string)")
	c.AlterExtentContainersAdd = NewCommandSymbol("AlterExtentContainersAdd", "(ExtentContainerId:guid, Url:string, State:string, CreatedOn:datetime, MaxDateTime:datetime, IsRecyclable:bool, StoresDatabaseMetadataPointer:bool, HardDeletePeriod:timespan, ActiveMetadataContainer:bool, MetadataContainer:bool)")
	c.AlterExtentContainersDrop = NewCommandSymbol("AlterExtentContainersDrop", "()")
	c.AlterExtentContainersRecycle = NewCommandSymbol("AlterExtentContainersRecycle", "()")
	c.AlterExtentContainersSet = NewCommandSymbol("AlterExtentContainersSet", "(ExtentContainerId:guid, Url:string, State:string, CreatedOn:datetime, MaxDateTime:datetime, IsRecyclable:bool, StoresDatabaseMetadataPointer:bool, HardDeletePeriod:timespan, ActiveMetadataContainer:bool, MetadataContainer:bool)")
	c.ShowExtentContainers = NewCommandSymbol("ShowExtentContainers", "(ExtentContainerId:guid, Url:string, State:string, CreatedOn:datetime, MaxDateTime:datetime, IsRecyclable:bool, StoresDatabaseMetadataPointer:bool, HardDeletePeriod:timespan, ActiveMetadataContainer:bool, MetadataContainer:bool)")
	c.DropEmptyExtentContainers = NewCommandSymbol("DropEmptyExtentContainers", "()")
	c.CleanDatabaseExtentContainers = NewCommandSymbol("CleanDatabaseExtentContainers", "()")
	c.ShowDatabaseExtentContainersCleanOperations = NewCommandSymbol("ShowDatabaseExtentContainersCleanOperations", "()")
	c.ClearDatabaseCacheQueryResults = NewCommandSymbol("ClearDatabaseCacheQueryResults", "()")
	c.ShowDatabaseCacheQueryResults = NewCommandSymbol("ShowDatabaseCacheQueryResults", "()")
	c.ShowDatabasesManagementGroups = NewCommandSymbol("ShowDatabasesManagementGroups", "()")
	c.AlterDatabaseStorageKeys = NewCommandSymbol("AlterDatabaseStorageKeys", "()")
	c.ClearDatabaseCacheStreamingIngestionSchema = NewCommandSymbol("ClearDatabaseCacheStreamingIngestionSchema", "()")
	c.ClearDatabaseCacheQueryWeakConsistency = NewCommandSymbol("ClearDatabaseCacheQueryWeakConsistency", "()")
	c.ShowEntitySchema = NewCommandSymbol("ShowEntitySchema", "()")
	c.ShowExtentDetails = NewCommandSymbol("ShowExtentDetails", "()")
	c.ShowExtentColumnStorageStats = NewCommandSymbol("ShowExtentColumnStorageStats", "()")
	c.AttachExtentsIntoTableByContainer = NewCommandSymbol("AttachExtentsIntoTableByContainer", "()")
	c.AttachExtentsIntoTableByMetadata = NewCommandSymbol("AttachExtentsIntoTableByMetadata", "()")
	c.AlterExtentTagsFromQuery = NewCommandSymbol("AlterExtentTagsFromQuery", "()")
	c.AlterMergeExtentTagsFromQuery = NewCommandSymbol("AlterMergeExtentTagsFromQuery", "()")
	c.DropExtentTagsFromQuery = NewCommandSymbol("DropExtentTagsFromQuery", "()")
	c.DropExtentTagsFromTable = NewCommandSymbol("DropExtentTagsFromTable", "()")
	c.DropExtentTagsRetention = NewCommandSymbol("DropExtentTagsRetention", "()")
	c.MergeDatabaseShardGroups = NewCommandSymbol("MergeDatabaseShardGroups", "()")
	c.AlterFollowerClusterConfiguration = NewCommandSymbol("AlterFollowerClusterConfiguration", "()")
	c.AddFollowerDatabaseAuthorizedPrincipals = NewCommandSymbol("AddFollowerDatabaseAuthorizedPrincipals", "()")
	c.DropFollowerDatabaseAuthorizedPrincipals = NewCommandSymbol("DropFollowerDatabaseAuthorizedPrincipals", "()")
	c.AlterFollowerDatabaseAuthorizedPrincipals = NewCommandSymbol("AlterFollowerDatabaseAuthorizedPrincipals", "()")
	c.DropFollowerDatabasePolicyCaching = NewCommandSymbol("DropFollowerDatabasePolicyCaching", "()")
	c.AlterFollowerDatabaseChildEntities = NewCommandSymbol("AlterFollowerDatabaseChildEntities", "()")
	c.AlterFollowerDatabaseConfiguration = NewCommandSymbol("AlterFollowerDatabaseConfiguration", "()")
	c.DropFollowerDatabases = NewCommandSymbol("DropFollowerDatabases", "()")
	c.ShowFollowerDatabase = NewCommandSymbol("ShowFollowerDatabase", "()")
	c.AlterFollowerTablesPolicyCaching = NewCommandSymbol("AlterFollowerTablesPolicyCaching", "()")
	c.DropFollowerTablesPolicyCaching = NewCommandSymbol("DropFollowerTablesPolicyCaching", "()")
	c.ShowFreshness = NewCommandSymbol("ShowFreshness", "()")
	c.ShowFunctionSchemaAsJson = NewCommandSymbol("ShowFunctionSchemaAsJson", "()")
	c.SetMaterializedViewAdmins = NewCommandSymbol("SetMaterializedViewAdmins", "()")
	c.AddMaterializedViewAdmins = NewCommandSymbol("AddMaterializedViewAdmins", "()")
	c.DropMaterializedViewAdmins = NewCommandSymbol("DropMaterializedViewAdmins", "()")
	c.SetMaterializedViewConcurrency = NewCommandSymbol("SetMaterializedViewConcurrency", "()")
	c.ClearMaterializedViewStatistics = NewCommandSymbol("ClearMaterializedViewStatistics", "()")
	c.ShowMaterializedViewStatistics = NewCommandSymbol("ShowMaterializedViewStatistics", "()")
	c.ShowMaterializedViewDiagnostics = NewCommandSymbol("ShowMaterializedViewDiagnostics", "()")
	c.ShowMaterializedViewFailures = NewCommandSymbol("ShowMaterializedViewFailures", "()")
	c.ShowMemory = NewCommandSymbol("ShowMemory", "()")
	c.CancelOperation = NewCommandSymbol("CancelOperation", "()")
	c.DisablePlugin = NewCommandSymbol("DisablePlugin", "()")
	c.EnablePlugin = NewCommandSymbol("EnablePlugin", "()")
	c.ShowPlugins = NewCommandSymbol("ShowPlugins", "()")
	c.ShowPrincipalAccess = NewCommandSymbol("ShowPrincipalAccess", "()")
	c.ShowDatabasePurgeOperation = NewCommandSymbol("ShowDatabasePurgeOperation", "()")
	c.ShowQueryExecution = NewCommandSymbol("ShowQueryExecution", "()")
	c.AlterPoliciesOfRetention = NewCommandSymbol("AlterPoliciesOfRetention", "()")
	c.DeletePoliciesOfRetention = NewCommandSymbol("DeletePoliciesOfRetention", "()")
	c.CreateRowStore = NewCommandSymbol("CreateRowStore", "()")
	c.DropRowStore = NewCommandSymbol("DropRowStore", "()")
	c.ShowRowStore = NewCommandSymbol("ShowRowStore", "(RowStoreName:string,RowStoreId:guid,RowStoreKey:string,OrdinalFrom:long,OrdinalTo:long,EstimatedDataSize:long,MinWriteAheadLogOffset:long, LocalStorageSize:long,LocalStorageStartOffset:long,Status:int,StatusLastUpdatedOn:datetime,DatabaseName:string,TableName:string,AssignedToNode:string,LatestIngestionTime:datetime)")
	c.ShowRowStores = NewCommandSymbol("ShowRowStores", "(RowStoreName:string,RowStoreId:guid,WriteAheadLogStorage:string,PersistentStorage:string,IsActive:bool,AssignedToNode:string,NumberOfKeys:long,WriteAheadLogSize:long, WriteAheadLogStartOffset:long,LocalStorageSize:long,WriteAheadDistanceToSizeRatioThreshold:real,InsertsConcurrencyLimit:long,KeyInsertsConcurrencyLimit:long,UnsealedSizePerKeyLimit:long, NodeInsertsConcurrencyLimit:long,Status:string,StatusLastUpdatedOn:datetime,UsageTags:string,IsEmpty:bool,IsDataAvailableForQuery:bool)")
	c.ShowRowStoreTransactions = NewCommandSymbol("ShowRowStoreTransactions", "()")
	c.ShowRowStoreSeals = NewCommandSymbol("ShowRowStoreSeals", "()")
	c.ShowSchema = NewCommandSymbol("ShowSchema", "()")
	c.ShowCallStacks = NewCommandSymbol("ShowCallStacks", "()")
	c.ShowFileSystem = NewCommandSymbol("ShowFileSystem", "()")
	c.ShowRunningCallouts = NewCommandSymbol("ShowRunningCallouts", "()")
	c.ShowStreamingIngestionFailures = NewCommandSymbol("ShowStreamingIngestionFailures", "(Database: string, Table: string, Principal: string, RootActivityId: guid, IngestionProperties: dynamic, Count: long, FirstFailureOn: datetime, LastFailureOn: datetime, FailureKind: string, ErrorCode: string, Details: string)")
	c.ShowStreamingIngestionStatistics = NewCommandSymbol("ShowStreamingIngestionStatistics", "(Database:string, Table:string, StartTime:datetime, EndTime:datetime, Count:long, MinDuration:timespan, MaxDuration:timespan, AvgDuration:timespan, TotalDataSize:long, MinDataSize:long, MaxDataSize:long, TotalRowCount:long, MinRowCount:long, MaxRowCount:long, IngestionStatus:string, NumOfRowStoresReferences:int, Principal:string, NodeId:string, IngestionProperties:dynamic)")
	c.AlterTableRowStoreReferencesDropKey = NewCommandSymbol("AlterTableRowStoreReferencesDropKey", "()")
	c.AlterTableRowStoreReferencesDropRowStore = NewCommandSymbol("AlterTableRowStoreReferencesDropRowStore", "()")
	c.AlterTableRowStoreReferencesDropBlockedKeys = NewCommandSymbol("AlterTableRowStoreReferencesDropBlockedKeys", "()")
	c.AlterTableRowStoreReferencesDisableKey = NewCommandSymbol("AlterTableRowStoreReferencesDisableKey", "()")
	c.AlterTableRowStoreReferencesDisableRowStore = NewCommandSymbol("AlterTableRowStoreReferencesDisableRowStore", "()")
	c.AlterTableRowStoreReferencesDisableBlockedKeys = NewCommandSymbol("AlterTableRowStoreReferencesDisableBlockedKeys", "()")
	c.SetTableRowStoreReferences = NewCommandSymbol("SetTableRowStoreReferences", "()")
	c.ShowTableRowStoreReferences = NewCommandSymbol("ShowTableRowStoreReferences", "(DatabaseName:string, TableName:string, RowstoreReferenceKey:string, RowStoreName:string, EnabledForIngestion:bool)")
	c.AlterTableColumnStatistics = NewCommandSymbol("AlterTableColumnStatistics", "()")
	c.AlterTableColumnStatisticsMethod = NewCommandSymbol("AlterTableColumnStatisticsMethod", "()")
	c.ShowTableColumnStatitics = NewCommandSymbol("ShowTableColumnStatitics", "()")
	c.ShowTableDimensions = NewCommandSymbol("ShowTableDimensions", "()")
	c.DeleteTableRecords = NewCommandSymbol("DeleteTableRecords", "()")
	c.TableDataUpdate = NewCommandSymbol("TableDataUpdate", "()")
	c.DeleteMaterializedViewRecords = NewCommandSymbol("DeleteMaterializedViewRecords", "()")
	c.ShowTableColumnsClassification = NewCommandSymbol("ShowTableColumnsClassification", "()")
	c.ShowTableRowStores = NewCommandSymbol("ShowTableRowStores", "(DatabaseName:string,TableName:string,ExtentId:guid,IsSealed:bool,RowStoreName:string,RowStoreId:string,RowStoreKey:string,OrdinalFrom:long,OrdinalTo:long,WriteAheadLogSize:long,LocalStorageSize:long,EstimatedDataSize:long,MinWriteAheadLogOffset:long)")
	c.ShowTableRowStoreSealInfo = NewCommandSymbol("ShowTableRowStoreSealInfo", "()")
	c.ShowTablesColumnStatistics = NewCommandSymbol("ShowTablesColumnStatistics", "()")
	c.ShowTableDataStatistics = NewCommandSymbol("ShowTableDataStatistics", "(ColumnName: string, ColumnType: string, ColumnId: guid, OriginalSize: long, ExtentSize: long, CompressionRatio: real, DataCompressionSize: long, SharedIndexSize: long, IndexSize: long, IndexSizePercent: real,StorageEngineVersion: string, PresentRowCount: long, DeletedRowCount: long, SamplePercent: real, IncludeColdData: bool)")
	c.CreateTempStorage = NewCommandSymbol("CreateTempStorage", "()")
	c.DropTempStorage = NewCommandSymbol("DropTempStorage", "()")
	c.DropStoredQueryResultContainers = NewCommandSymbol("DropStoredQueryResultContainers", "()")
	c.DropUnusedStoredQueryResultContainers = NewCommandSymbol("DropUnusedStoredQueryResultContainers", "()")
	c.EnableDatabaseMaintenanceMode = NewCommandSymbol("EnableDatabaseMaintenanceMode", "()")
	c.DisableDatabaseMaintenanceMode = NewCommandSymbol("DisableDatabaseMaintenanceMode", "()")
	c.EnableDatabaseStreamingIngestionMaintenanceMode = NewCommandSymbol("EnableDatabaseStreamingIngestionMaintenanceMode", "()")
	c.DisableDatabaseStreamingIngestionMaintenanceMode = NewCommandSymbol("DisableDatabaseStreamingIngestionMaintenanceMode", "()")
	c.ShowQueryCallTree = NewCommandSymbol("ShowQueryCallTree", "()")
	c.ShowExtentCorruptedDatetime = NewCommandSymbol("ShowExtentCorruptedDatetime", "()")
	c.PatchExtentCorruptedDatetime = NewCommandSymbol("PatchExtentCorruptedDatetime", "()")
	c.ClearClusterCredStoreCache = NewCommandSymbol("ClearClusterCredStoreCache", "()")
	c.ClearClusterOutboundAccessPolicyCache = NewCommandSymbol("ClearClusterOutboundAccessPolicyCache", "()")
	c.ClearClusterGroupMembershipCache = NewCommandSymbol("ClearClusterGroupMembershipCache", "()")
	c.ClearExternalArtifactsCache = NewCommandSymbol("ClearExternalArtifactsCache", "()")
	c.ShowDatabasesEntities = NewCommandSymbol("ShowDatabasesEntities", "(DatabaseName:string, EntityType:string, EntityName:string, DocString:string, Folder:string, CslInputSchema:string, Content:string, CslOutputSchema:string, Properties:dynamic)")
	c.ShowDatabaseEntity = NewCommandSymbol("ShowDatabaseEntity", "(DatabaseName:string, EntityType:string, EntityName:string, DocString:string, Folder:string, CslInputSchema:string, Content:string, CslOutputSchema:string, Properties:dynamic)")
	c.ReplaceDatabaseKeyVaultSecrets = NewCommandSymbol("ReplaceDatabaseKeyVaultSecrets", "()")
	c.All = []*CommandSymbol{
		c.ShowDatabase,
		c.ShowDatabaseDetails,
		c.ShowDatabaseIdentity,
		c.ShowDatabasePolicies,
		c.ShowDatabaseDataStats,
		c.ShowDatabaseMetadata,
		c.ShowClusterDatabases,
		c.ShowClusterDatabasesDetails,
		c.ShowClusterDatabasesIdentity,
		c.ShowClusterDatabasesPolicies,
		c.ShowClusterDatabasesDataStats,
		c.ShowClusterDatabasesMetadata,
		c.CreateDatabase,
		c.AttachDatabase,
		c.DetachDatabase,
		c.AlterDatabasePrettyName,
		c.DropDatabasePrettyName,
		c.AlterDatabasePersistMetadata,
		c.SetAccess,
		c.ShowDatabaseSchema,
		c.ShowDatabaseSchemaAsJson,
		c.ShowDatabaseSchemaAsCslScript,
		c.ShowDatabaseCslSchema,
		c.ShowDatabaseSchemaViolations,
		c.ShowDatabasesSchema,
		c.ShowDatabasesSchemaAsJson,
		c.CreateDatabaseIngestionMapping,
		c.CreateOrAlterDatabaseIngestionMapping,
		c.AlterDatabaseIngestionMapping,
		c.AlterMergeDatabaseIngestionMapping,
		c.ShowDatabaseIngestionMappings,
		c.ShowIngestionMappings,
		c.DropDatabaseIngestionMapping,
		c.ShowTables,
		c.ShowTable,
		c.ShowTablesDetails,
		c.ShowTableDetails,
		c.ShowTableCslSchema,
		c.ShowTableSchemaAsJson,
		c.CreateTable,
		c.CreateTableBasedOnAnother,
		c.CreateMergeTable,
		c.DefineTable,
		c.CreateTables,
		c.CreateMergeTables,
		c.DefineTables,
		c.AlterTable,
		c.AlterMergeTable,
		c.AlterTableDocString,
		c.AlterTableFolder,
		c.RenameTable,
		c.RenameTables,
		c.UndoDropExtentContainer,
		c.DropTable,
		c.UndoDropTable,
		c.DropTables,
		c.CreateTableIngestionMapping,
		c.CreateOrAlterTableIngestionMapping,
		c.AlterTableIngestionMapping,
		c.AlterMergeTableIngestionMapping,
		c.ShowTableIngestionMappings,
		c.ShowTableIngestionMapping,
		c.DropTableIngestionMapping,
		c.RenameColumn,
		c.RenameColumns,
		c.AlterColumnType,
		c.DropColumn,
		c.DropTableColumns,
		c.AlterTableColumnDocStrings,
		c.AlterMergeTableColumnDocStrings,
		c.ShowFunctions,
		c.ShowFunction,
		c.CreateFunction,
		c.AlterFunction,
		c.CreateOrAlterFunction,
		c.DropFunction,
		c.DropFunctions,
		c.AlterFunctionDocString,
		c.AlterFunctionFolder,
		c.ShowExternalTables,
		c.ShowExternalTable,
		c.ShowExternalTablesDetails,
		c.ShowExternalTableDetails,
		c.ShowExternalTableCslSchema,
		c.ShowExternalTableSchema,
		c.ShowExternalTableArtifacts,
		c.DropExternalTable,
		c.CreateStorageExternalTable,
		c.AlterStorageExternalTable,
		c.CreateOrAlterStorageExternalTable,
		c.CreateSqlExternalTable,
		c.AlterSqlExternalTable,
		c.CreateOrAlterSqlExternalTable,
		c.CreateExternalTableMapping,
		c.SetExternalTableAdmins,
		c.AddExternalTableAdmins,
		c.DropExternalTableAdmins,
		c.AlterExternalTableDocString,
		c.AlterExternalTableFolder,
		c.ShowExternalTablePrincipals,
		c.ShowFabric,
		c.AlterExternalTableMapping,
		c.ShowExternalTableMappings,
		c.ShowExternalTableMapping,
		c.DropExternalTableMapping,
		c.ShowWorkloadGroups,
		c.ShowWorkloadGroup,
		c.CreateOrAleterWorkloadGroup,
		c.AlterMergeWorkloadGroup,
		c.DropWorkloadGroup,
		c.ShowDatabasePolicyCaching,
		c.ShowTablePolicyCaching,
		c.ShowTableStarPolicyCaching,
		c.ShowColumnPolicyCaching,
		c.ShowMaterializedViewPolicyCaching,
		c.ShowGraphModelPolicyCaching,
		c.ShowGraphModelStarPolicyCaching,
		c.ShowClusterPolicyCaching,
		c.AlterDatabasePolicyCaching,
		c.AlterTablePolicyCaching,
		c.AlterTablesPolicyCaching,
		c.AlterColumnPolicyCaching,
		c.AlterMaterializedViewPolicyCaching,
		c.AlterGraphModelPolicyCaching,
		c.AlterClusterPolicyCaching,
		c.DeleteDatabasePolicyCaching,
		c.DeleteTablePolicyCaching,
		c.DeleteColumnPolicyCaching,
		c.DeleteMaterializedViewPolicyCaching,
		c.DeleteGraphModelPolicyCaching,
		c.DeleteClusterPolicyCaching,
		c.ShowTablePolicyChangeDataCapture,
		c.AlterTablePolicyChangeDataCapture,
		c.AlterTablesPolicyChangeDataCapture,
		c.DeleteTablePolicyChangeDataCapture,
		c.ShowTablePolicyIngestionTime,
		c.ShowTableStarPolicyIngestionTime,
		c.AlterTablePolicyIngestionTime,
		c.AlterTablesPolicyIngestionTime,
		c.DeleteTablePolicyIngestionTime,
		c.ShowDatabasePolicyRowId,
		c.AlterDatabasePolicyRowId,
		c.DeleteDatabasePolicyRowId,
		c.ShowTablePolicyRowId,
		c.AlterTablePolicyRowId,
		c.AlterTablesPolicyRowId,
		c.DeleteTablePolicyRowId,
		c.ShowTablePolicyRetention,
		c.ShowTableStarPolicyRetention,
		c.ShowGraphPolicyRetention,
		c.ShowGraphStarPolicyRetention,
		c.ShowDatabasePolicyRetention,
		c.AlterTablePolicyRetention,
		c.AlterMaterializedViewPolicyRetention,
		c.AlterDatabasePolicyRetention,
		c.AlterGraphModelPolicyRetention,
		c.AlterTablesPolicyRetention,
		c.AlterMergeTablePolicyRetention,
		c.AlterMergeMaterializedViewPolicyRetention,
		c.AlterMergeDatabasePolicyRetention,
		c.DeleteTablePolicyRetention,
		c.DeleteDatabasePolicyRetention,
		c.ShowDatabasePolicyHardRetentionViolations,
		c.ShowDatabasePolicySoftRetentionViolations,
		c.ShowTablePolicyRowLevelSecurity,
		c.ShowTableStarPolicyRowLevelSecurity,
		c.AlterTablePolicyRowLevelSecurity,
		c.DeleteTablePolicyRowLevelSecurity,
		c.ShowMaterializedViewPolicyRowLevelSecurity,
		c.AlterMaterializedViewPolicyRowLevelSecurity,
		c.DeleteMaterializedViewPolicyRowLevelSecurity,
		c.ShowExternalTablePolicyRowLevelSecurity,
		c.ShowExternalTableStarPolicyRowLevelSecurity,
		c.AlterExternalTablePolicyRowLevelSecurity,
		c.DeleteExternalTablePolicyRowLevelSecurity,
		c.ShowTablePolicyRowOrder,
		c.ShowTableStarPolicyRowOrder,
		c.AlterTablePolicyRowOrder,
		c.AlterTablesPolicyRowOrder,
		c.AlterMergeTablePolicyRowOrder,
		c.DeleteTablePolicyRowOrder,
		c.ShowTablePolicyUpdate,
		c.ShowTableStarPolicyUpdate,
		c.AlterTablePolicyUpdate,
		c.AlterMergeTablePolicyUpdate,
		c.DeleteTablePolicyUpdate,
		c.ShowClusterPolicyIngestionBatching,
		c.ShowDatabasePolicyIngestionBatching,
		c.ShowTablePolicyIngestionBatching,
		c.ShowTableStarPolicyIngestionBatching,
		c.AlterClusterPolicyIngestionBatching,
		c.AlterMergeClusterPolicyIngestionBatching,
		c.AlterDatabasePolicyIngestionBatching,
		c.AlterMergeDatabasePolicyIngestionBatching,
		c.AlterTablePolicyIngestionBatching,
		c.AlterMergeTablePolicyIngestionBatching,
		c.AlterTablesPolicyIngestionBatching,
		c.DeleteDatabasePolicyIngestionBatching,
		c.DeleteTablePolicyIngestionBatching,
		c.ShowDatabasePolicyEncoding,
		c.ShowTablePolicyEncoding,
		c.ShowColumnPolicyEncoding,
		c.AlterDatabasePolicyEncoding,
		c.AlterTablePolicyEncoding,
		c.AlterTableColumnsPolicyEncoding,
		c.AlterColumnPolicyEncoding,
		c.AlterColumnsPolicyEncodingByQuery,
		c.AlterColumnPolicyEncodingType,
		c.AlterMergeDatabasePolicyEncoding,
		c.AlterMergeTablePolicyEncoding,
		c.AlterMergeColumnPolicyEncoding,
		c.DeleteDatabasePolicyEncoding,
		c.DeleteTablePolicyEncoding,
		c.DeleteColumnPolicyEncoding,
		c.ShowDatabasePolicyMerge,
		c.ShowTablePolicyMerge,
		c.ShowTableStarPolicyMerge,
		c.AlterDatabasePolicyMerge,
		c.AlterTablePolicyMerge,
		c.AlterTablesPolicyMerge,
		c.AlterMergeDatabasePolicyMerge,
		c.AlterMergeTablePolicyMerge,
		c.AlterMergeMaterializedViewPolicyMerge,
		c.DeleteDatabasePolicyMerge,
		c.DeleteTablePolicyMerge,
		c.ShowExternalTablePolicyQueryAcceleration,
		c.ShowExternalTablesPolicyQueryAcceleration,
		c.AlterExternalTablePolicyQueryAcceleration,
		c.AlterMergeExternalTablePolicyQueryAcceleration,
		c.DeleteExternalTablePolicyQueryAcceleration,
		c.ShowExternalTableQueryAccelerationStatatistics,
		c.ShowExternalTablesQueryAccelerationStatatistics,
		c.AlterTablePolicyMirroring,
		c.AlterMergeTablePolicyMirroring,
		c.AlterTablePolicyMirroringWithJson,
		c.AlterMergeTablePolicyMirroringWithJson,
		c.DeleteTablePolicyMirroring,
		c.ShowTablePolicyMirroring,
		c.ShowTableStarPolicyMirroring,
		c.CreateMirroringTemplate,
		c.CreateOrAlterMirroringTemplate,
		c.AlterMirroringTemplate,
		c.AlterMergeMirroringTemplate,
		c.DeleteMirroringTemplate,
		c.ShowMirroringTemplate,
		c.ShowMirroringTemplates,
		c.ApplyMirroringTemplate,
		c.ShowTablePolicyPartitioning,
		c.ShowTableStarPolicyPartitioning,
		c.AlterTablePolicyPartitioning,
		c.AlterMergeTablePolicyPartitioning,
		c.AlterMaterializedViewPolicyPartitioning,
		c.AlterMergeMaterializedViewPolicyPartitioning,
		c.DeleteTablePolicyPartitioning,
		c.DeleteMaterializedViewPolicyPartitioning,
		c.ShowTablePolicyRestrictedViewAccess,
		c.ShowTableStarPolicyRestrictedViewAccess,
		c.AlterTablePolicyRestrictedViewAccess,
		c.AlterTablesPolicyRestrictedViewAccess,
		c.DeleteTablePolicyRestrictedViewAccess,
		c.ShowClusterPolicyRowStore,
		c.AlterClusterPolicyRowStore,
		c.AlterMergeClusterPolicyRowStore,
		c.DeleteClusterPolicyRowStore,
		c.ShowClusterPolicySandbox,
		c.AlterClusterPolicySandbox,
		c.DeleteClusterPolicySandbox,
		c.ShowClusterSandboxesStats,
		c.ShowDatabasePolicySharding,
		c.ShowTablePolicySharding,
		c.ShowTableStarPolicySharding,
		c.AlterDatabasePolicySharding,
		c.AlterTablePolicySharding,
		c.AlterMergeDatabasePolicySharding,
		c.AlterMergeTablePolicySharding,
		c.DeleteDatabasePolicySharding,
		c.DeleteTablePolicySharding,
		c.AlterClusterPolicySharding,
		c.AlterMergeClusterPolicySharding,
		c.DeleteClusterPolicySharding,
		c.ShowClusterPolicySharding,
		c.ShowDatabasePolicyShardsGrouping,
		c.AlterDatabasePolicyShardsGrouping,
		c.AlterMergeDatabasePolicyShardsGrouping,
		c.DeleteDatabasePolicyShardsGrouping,
		c.ShowDatabasePolicyStreamingIngestion,
		c.ShowTablePolicyStreamingIngestion,
		c.ShowClusterPolicyStreamingIngestion,
		c.AlterDatabasePolicyStreamingIngestion,
		c.AlterMergeDatabasePolicyStreamingIngestion,
		c.AlterTablePolicyStreamingIngestion,
		c.AlterMergeTablePolicyStreamingIngestion,
		c.AlterClusterPolicyStreamingIngestion,
		c.AlterMergeClusterPolicyStreamingIngestion,
		c.DeleteDatabasePolicyStreamingIngestion,
		c.DeleteTablePolicyStreamingIngestion,
		c.DeleteClusterPolicyStreamingIngestion,
		c.ShowDatabasePolicyManagedIdentity,
		c.ShowClusterPolicyManagedIdentity,
		c.AlterDatabasePolicyManagedIdentity,
		c.AlterMergeDatabasePolicyManagedIdentity,
		c.AlterClusterPolicyManagedIdentity,
		c.AlterMergeClusterPolicyManagedIdentity,
		c.DeleteDatabasePolicyManagedIdentity,
		c.DeleteClusterPolicyManagedIdentity,
		c.ShowTablePolicyAutoDelete,
		c.AlterTablePolicyAutoDelete,
		c.DeleteTablePolicyAutoDelete,
		c.ShowClusterPolicyCallout,
		c.AlterClusterPolicyCallout,
		c.AlterMergeClusterPolicyCallout,
		c.DeleteClusterPolicyCallout,
		c.ShowClusterPolicyCapacity,
		c.AlterClusterPolicyCapacity,
		c.AlterMergeClusterPolicyCapacity,
		c.ShowClusterPolicyMaterializedViewPrioritization,
		c.AlterClusterPolicyMaterializedViewPrioritization,
		c.AlterMergeClusterPolicyMaterializedViewPrioritization,
		c.DropClusterPolicyMaterializedViewPrioritization,
		c.ShowClusterPolicyRequestClassification,
		c.AlterClusterPolicyRequestClassification,
		c.AlterMergeClusterPolicyRequestClassification,
		c.DeleteClusterPolicyRequestClassification,
		c.ShowClusterPolicyMultiDatabaseAdmins,
		c.AlterClusterPolicyMultiDatabaseAdmins,
		c.AlterMergeClusterPolicyMultiDatabaseAdmins,
		c.ShowDatabasePolicyDiagnostics,
		c.ShowClusterPolicyDiagnostics,
		c.AlterDatabasePolicyDiagnostics,
		c.AlterMergeDatabasePolicyDiagnostics,
		c.AlterClusterPolicyDiagnostics,
		c.AlterMergeClusterPolicyDiagnostics,
		c.DeleteDatabasePolicyDiagnostics,
		c.ShowClusterPolicyQueryWeakConsistency,
		c.AlterClusterPolicyQueryWeakConsistency,
		c.AlterMergeClusterPolicyQueryWeakConsistency,
		c.ShowTablePolicyExtentTagsRetention,
		c.ShowTableStarPolicyExtentTagsRetention,
		c.ShowDatabasePolicyExtentTagsRetention,
		c.AlterTablePolicyExtentTagsRetention,
		c.AlterDatabasePolicyExtentTagsRetention,
		c.DeleteTablePolicyExtentTagsRetention,
		c.DeleteDatabasePolicyExtentTagsRetention,
		c.ShowPrincipalRoles,
		c.ShowDatabasePrincipalRoles,
		c.ShowTablePrincipalRoles,
		c.ShowGraphModelPrincipalRoles,
		c.ShowExternalTablesPrincipalRoles,
		c.ShowFunctionPrincipalRoles,
		c.ShowClusterPrincipalRoles,
		c.ShowClusterPrincipals,
		c.ShowDatabasePrincipals,
		c.ShowTablePrincipals,
		c.ShowGraphModelPrincipals,
		c.ShowFunctionPrincipals,
		c.AddClusterRole,
		c.DropClusterRole,
		c.SetClusterRole,
		c.AddDatabaseRole,
		c.DropDatabaseRole,
		c.SetDatabaseRole,
		c.AddTableRole,
		c.DropTableRole,
		c.SetTableRole,
		c.AddFunctionRole,
		c.DropFunctionRole,
		c.SetFunctionRole,
		c.ShowClusterBlockedPrincipals,
		c.AddClusterBlockedPrincipals,
		c.DropClusterBlockedPrincipals,
		c.SetClusterMaintenanceMode,
		c.IngestIntoTable,
		c.IngestInlineIntoTable,
		c.SetTable,
		c.AppendTable,
		c.SetOrAppendTable,
		c.SetOrReplaceTable,
		c.ExportToStorage,
		c.ExportToSqlTable,
		c.ExportToExternalTable,
		c.CreateOrAlterContinuousExport,
		c.ShowContinuousExport,
		c.ShowContinuousExports,
		c.ShowClusterPendingContinuousExports,
		c.ShowContinuousExportExportedArtifacts,
		c.ShowContinuousExportFailures,
		c.SetContinuousExportCursor,
		c.DropContinuousExport,
		c.EnableContinuousExport,
		c.DisableContinuousExport,
		c.CreateMaterializedView,
		c.CreateMaterializedViewOverMaterializedView,
		c.RenameMaterializedView,
		c.ShowMaterializedView,
		c.ShowMaterializedViews,
		c.ShowMaterializedViewsDetails,
		c.ShowMaterializedViewDetails,
		c.ShowMaterializedViewPolicyRetention,
		c.ShowMaterializedViewPolicyMerge,
		c.ShowMaterializedViewPolicyPartitioning,
		c.ShowMaterializedViewExtents,
		c.AlterMaterializedView,
		c.AlterMaterializedViewOverMaterializedView,
		c.CreateOrAlterMaterializedView,
		c.CreateOrAlterMaterializedViewOverMaterializedView,
		c.DropMaterializedView,
		c.EnableDisableMaterializedView,
		c.ShowMaterializedViewPrincipals,
		c.ShowMaterializedViewSchemaAsJson,
		c.ShowMaterializedViewCslSchema,
		c.AlterMaterializedViewFolder,
		c.AlterMaterializedViewDocString,
		c.AlterMaterializedViewLookback,
		c.AlterMaterializedViewAutoUpdateSchema,
		c.ClearMaterializedViewData,
		c.SetMaterializedViewCursor,
		c.ShowTableOperationsMirroringStatus,
		c.ShowTableOperationsMirroringExportedArtifacts,
		c.ShowTableOperationsMirroringFailures,
		c.ShowCluster,
		c.ShowClusterDetails,
		c.ShowDiagnostics,
		c.ShowCapacity,
		c.ShowOperations,
		c.ShowOperationDetails,
		c.ShowJournal,
		c.ShowDatabaseJournal,
		c.ShowClusterJournal,
		c.ShowQueries,
		c.ShowRunningQueries,
		c.CancelQuery,
		c.ShowQueryPlan,
		c.ShowCache,
		c.AlterCache,
		c.ShowCommands,
		c.ShowCommandsAndQueries,
		c.ShowIngestionFailures,
		c.ShowDataOperations,
		c.ShowDatabaseKeyVaultSecrets,
		c.ShowClusterExtents,
		c.ShowClusterExtentsMetadata,
		c.ShowDatabaseExtents,
		c.ShowDatabaseExtentsMetadata,
		c.ShowDatabaseExtentTagsStatistics,
		c.ShowDatabaseExtentsPartitioningStatistics,
		c.ShowTableExtents,
		c.ShowTableExtentsMetadata,
		c.TableShardsGroupShow,
		c.TableShardsGroupMetadataShow,
		c.TableShardGroupsShow,
		c.TableShardGroupsStatisticsShow,
		c.TablesShardGroupsStatisticsShow,
		c.DatabaseShardGroupsStatisticsShow,
		c.MergeExtents,
		c.MergeExtentsDryrun,
		c.MoveExtentsFrom,
		c.MoveExtentsQuery,
		c.TableShuffleExtents,
		c.TableShuffleExtentsQuery,
		c.ReplaceExtents,
		c.DropExtent,
		c.DropExtents,
		c.DropExtentsPartitionMetadata,
		c.DropPretendExtentsByProperties,
		c.ShowVersion,
		c.ClearTableData,
		c.ClearTableCacheStreamingIngestionSchema,
		c.ShowStorageArtifactsCleanupState,
		c.ClusterDropStorageArtifactsCleanupState,
		c.StoredQueryResultSet,
		c.StoredQueryResultSetOrReplace,
		c.StoredQueryResultsShow,
		c.StoredQueryResultShowSchema,
		c.StoredQueryResultDrop,
		c.StoredQueryResultsDrop,
		c.GraphModelCreateOrAlter,
		c.GraphModelShow,
		c.GraphModelsShow,
		c.GraphModelDrop,
		c.SetGraphModelAdmins,
		c.AddGraphModelAdmins,
		c.DropGraphModelAdmins,
		c.GraphSnapshotMake,
		c.GraphSnapshotShow,
		c.GraphSnapshotsShow,
		c.GraphSnapshotDrop,
		c.GraphSnapshotsDrop,
		c.GraphSnapshotShowDataStatistics,
		c.GraphSnapshotsShowDataStatistics,
		c.GraphSnapshotShowStatistics,
		c.GraphSnapshotsShowStatistics,
		c.GraphSnapshotShowFailures,
		c.GraphShardMake,
		c.GraphShardDrop,
		c.ShowCertificates,
		c.ShowClockDiagnostics,
		c.ShowCloudSettings,
		c.ShowCommConcurrency,
		c.ShowCommPools,
		c.ShowFabricCache,
		c.ShowFabricLocks,
		c.ShowFabricClocks,
		c.ShowFeatureFlags,
		c.ShowMemPools,
		c.ShowServicePoints,
		c.ShowTcpConnections,
		c.ShowTcpPorts,
		c.ShowThreadPools,
		c.ShowWhatsGoingOn,
		c.ExecuteDatabaseScript,
		c.ExecuteClusterScript,
		c.CreateRequestMemoryProfile,
		c.CreateRequestSupport,
		c.ShowRequestSupport,
		c.ShowClusterAdminState,
		c.ClearRemoteClusterDatabaseSchema,
		c.ShowClusterMonitoring,
		c.ShowClusterScaleIn,
		c.ShowClusterServices,
		c.ShowClusterNetwork,
		c.AlterClusterStorageKeys,
		c.ShowClusterStorageKeysHash,
		c.AlterFabricServiceAssignmentsCommand,
		c.DropFabricServiceAssignmentsCommand,
		c.CreateEntityGroupCommand,
		c.CreateOrAlterEntityGroupCommand,
		c.AlterEntityGroup,
		c.AlterMergeEntityGroup,
		c.DropEntityGroup,
		c.ShowEntityGroup,
		c.ShowEntityGroups,
		c.AlterExtentContainersAdd,
		c.AlterExtentContainersDrop,
		c.AlterExtentContainersRecycle,
		c.AlterExtentContainersSet,
		c.ShowExtentContainers,
		c.DropEmptyExtentContainers,
		c.CleanDatabaseExtentContainers,
		c.ShowDatabaseExtentContainersCleanOperations,
		c.ClearDatabaseCacheQueryResults,
		c.ShowDatabaseCacheQueryResults,
		c.ShowDatabasesManagementGroups,
		c.AlterDatabaseStorageKeys,
		c.ClearDatabaseCacheStreamingIngestionSchema,
		c.ClearDatabaseCacheQueryWeakConsistency,
		c.ShowEntitySchema,
		c.ShowExtentDetails,
		c.ShowExtentColumnStorageStats,
		c.AttachExtentsIntoTableByContainer,
		c.AttachExtentsIntoTableByMetadata,
		c.AlterExtentTagsFromQuery,
		c.AlterMergeExtentTagsFromQuery,
		c.DropExtentTagsFromQuery,
		c.DropExtentTagsFromTable,
		c.DropExtentTagsRetention,
		c.MergeDatabaseShardGroups,
		c.AlterFollowerClusterConfiguration,
		c.AddFollowerDatabaseAuthorizedPrincipals,
		c.DropFollowerDatabaseAuthorizedPrincipals,
		c.AlterFollowerDatabaseAuthorizedPrincipals,
		c.DropFollowerDatabasePolicyCaching,
		c.AlterFollowerDatabaseChildEntities,
		c.AlterFollowerDatabaseConfiguration,
		c.DropFollowerDatabases,
		c.ShowFollowerDatabase,
		c.AlterFollowerTablesPolicyCaching,
		c.DropFollowerTablesPolicyCaching,
		c.ShowFreshness,
		c.ShowFunctionSchemaAsJson,
		c.SetMaterializedViewAdmins,
		c.AddMaterializedViewAdmins,
		c.DropMaterializedViewAdmins,
		c.SetMaterializedViewConcurrency,
		c.ClearMaterializedViewStatistics,
		c.ShowMaterializedViewStatistics,
		c.ShowMaterializedViewDiagnostics,
		c.ShowMaterializedViewFailures,
		c.ShowMemory,
		c.CancelOperation,
		c.DisablePlugin,
		c.EnablePlugin,
		c.ShowPlugins,
		c.ShowPrincipalAccess,
		c.ShowDatabasePurgeOperation,
		c.ShowQueryExecution,
		c.AlterPoliciesOfRetention,
		c.DeletePoliciesOfRetention,
		c.CreateRowStore,
		c.DropRowStore,
		c.ShowRowStore,
		c.ShowRowStores,
		c.ShowRowStoreTransactions,
		c.ShowRowStoreSeals,
		c.ShowSchema,
		c.ShowCallStacks,
		c.ShowFileSystem,
		c.ShowRunningCallouts,
		c.ShowStreamingIngestionFailures,
		c.ShowStreamingIngestionStatistics,
		c.AlterTableRowStoreReferencesDropKey,
		c.AlterTableRowStoreReferencesDropRowStore,
		c.AlterTableRowStoreReferencesDropBlockedKeys,
		c.AlterTableRowStoreReferencesDisableKey,
		c.AlterTableRowStoreReferencesDisableRowStore,
		c.AlterTableRowStoreReferencesDisableBlockedKeys,
		c.SetTableRowStoreReferences,
		c.ShowTableRowStoreReferences,
		c.AlterTableColumnStatistics,
		c.AlterTableColumnStatisticsMethod,
		c.ShowTableColumnStatitics,
		c.ShowTableDimensions,
		c.DeleteTableRecords,
		c.TableDataUpdate,
		c.DeleteMaterializedViewRecords,
		c.ShowTableColumnsClassification,
		c.ShowTableRowStores,
		c.ShowTableRowStoreSealInfo,
		c.ShowTablesColumnStatistics,
		c.ShowTableDataStatistics,
		c.CreateTempStorage,
		c.DropTempStorage,
		c.DropStoredQueryResultContainers,
		c.DropUnusedStoredQueryResultContainers,
		c.EnableDatabaseMaintenanceMode,
		c.DisableDatabaseMaintenanceMode,
		c.EnableDatabaseStreamingIngestionMaintenanceMode,
		c.DisableDatabaseStreamingIngestionMaintenanceMode,
		c.ShowQueryCallTree,
		c.ShowExtentCorruptedDatetime,
		c.PatchExtentCorruptedDatetime,
		c.ClearClusterCredStoreCache,
		c.ClearClusterOutboundAccessPolicyCache,
		c.ClearClusterGroupMembershipCache,
		c.ClearExternalArtifactsCache,
		c.ShowDatabasesEntities,
		c.ShowDatabaseEntity,
		c.ReplaceDatabaseKeyVaultSecrets,
	}
	return c
}
