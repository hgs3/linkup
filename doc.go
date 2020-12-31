// LinkUp - A tool for catching broken website links.
// Copyright (C) 2020-2021 Henry G. Stratmann III
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

/*
Package linkup detects broken links on your website.
It works by traversing your sites HTML documents and verifies all links connect to valid files or folders.

It can verify internal or external links:
An internal link connects web pages on the same domain and an external link connects to a web page on a separate domain.
External links are verified by pinging them to ensure validity.
*/
package linkup
