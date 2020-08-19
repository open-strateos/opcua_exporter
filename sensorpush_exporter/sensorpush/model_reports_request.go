/*
 * SensorPush Public API
 *
 * This is a swagger definition for the SensorPush public REST API. Download the definition file [here](https://api.sensorpush.com/api/v1/support/swagger/swagger-v1.json).
 *
 * API version: v1.0.20200621
 * Contact: support@sensorpush.com
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package sensorpush
// ReportsRequest Request object for reports.
type ReportsRequest struct {
	// The directory path to perform this operation.
	Path string `json:"path,omitempty"`
}
