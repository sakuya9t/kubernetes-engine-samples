// Package query defines MQL queries used for cloud monitoring service.
package query

import (
	"consumptionexp/config"
	"fmt"
)

// CPUUsageTime is the query of the metrics kubernetes.io/container/cpu/core_usage_time.
var CPUUsageTime = fmt.Sprintf(`fetch k8s_container
																	| metric 'kubernetes.io/container/cpu/core_usage_time'
																	| align rate(%s)
																	| every %s
																	| group_by
																			[resource.project_id, resource.location, resource.cluster_name,
																			resource.namespace_name, resource.pod_name, resource.container_name],
																			[value_core_usage_time_aggregate: aggregate(value.core_usage_time)]
																	| within %s, %s`,
	config.DefaultResolution, config.DefaultResolution, config.DefaultStart, config.DefaultPeriod)

// MemoryUsedBytes is the query of the metrics kubernetes.io/container/memory/used_bytes.
var MemoryUsedBytes = fmt.Sprintf(`fetch k8s_container
																	| metric 'kubernetes.io/container/memory/used_bytes'
																	| group_by %s, [value_used_bytes_mean: mean(value.used_bytes)]
																	| every %s
																	| group_by
																			[resource.project_id, resource.location, resource.cluster_name,
																			resource.namespace_name, resource.pod_name, resource.container_name],
																			[value_used_bytes_mean_mean: mean(value_used_bytes_mean)]
																	| within %s, %s`,
	config.DefaultResolution, config.DefaultResolution, config.DefaultStart, config.DefaultPeriod)
