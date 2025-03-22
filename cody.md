```
Let's start a new project.  We're going to use Go to write an inventory checker for the MicroCenter website.  This program will run as a service and periodically check the MicroCenter website for a specfic store id and a list of produt pages.   We'll use the go-selenium library as a headless browser to load the pages and find the information about the number of items each product has in stock at that store.  We will export prometheus metrics with gauges for the current count of each product at the store. 
```

```
It's going to take some trial and error to get this to work with MicroCenter's actual website structure.  Let's create a test for checkProductInventory so that we can iterate on this without needing to run the entire program.  Also, let's decouple checkProductInventory from the prometheus metrics.  checkProductInventory can just return the data (so that we can validate it from the tests) and the main program can update the prometheus gauge with the result of checkProductInventory
```