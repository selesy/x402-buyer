// Package buyer provides a library of code that allows the standard
// library's http.Client to pay for HTTP content and services using the
// x402 protocol.
//
// It is anticipated that this software will commonly be used to allow
// AI agents to pay for the services they need.  When allowing automated
// payments on your behalf, care should be taken to limit your financial
// exposure.
//
// Defaults
//
//   - If the WithClient option is not specified, the http.DefaultClient
//     is used with the http.DefaultTransport.
//   - If the WithLogger Option is not specified, a No-Op logger is used.
package buyer
