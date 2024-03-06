# soqlnowhere

Find SOQL queries in Apex classes and triggers that contain no WHERE clause.

This is an example of using the antlr parser from
[apexfmt](https://github.com/octoberswimmer/apexfmt) to identify queries in
Apex files that have no WHERE clause.

It uses a listener to check each SOQL query for a WHERE clause, then uses the
formatter in apexfmt to output the formatted query.
