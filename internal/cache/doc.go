// Package cache manages the disposable cache directory under the XDG cache path.
//
// Live cache entries are immediate children of the cache directory. Names containing
// ".corrupt." or ".partial." are preserved aside data from recovery or atomic writes,
// not live cache. Future cache readers must skip those infixes.
package cache
