# golang

Storage problem

We receive some records in a CSV file (example promotions.csv attached) every 30
minutes. We would like to store these objects in a way to be accessed by an endpoint.
Given an ID the endpoint should return the object, otherwise, return not found.

Eg:
curl https://localhost:1321/promotions/1

{"id":"172FFC14-D229-4C93-B06B-F48B8C095512", "price":9.68,
"expiration_date": "2022-06-04 06:01:20"}


Additionally, consider:
1) The .csv file could be very big (billions of entries) - how would your application
perform? 
2) Every new file is immutable, that is, you should erase and write the whole storage;
3) How would your application perform in peak periods (millions of requests per
minute)?
4) How would you operate this app in production (e.g. deployment, scaling, monitoring)?
5) The application should be written in golang;
6) Main deliverable is the code for the app including usage instructions, ideally in a
repo/github gist.


Answers
-----------------------------------------------------------------------------
1)To efficiently store a large amount of data, it is recommended to utilize concurrency along with a pool of worker jobs to insert the data into a MySQL database. The number of workers can be controlled based on the size of the data being processed.
2) The program simply reads the data from the csv file and does not change it.
3) The main issue is data storage, so the pooling method can help improve application performance.
4) For Deployment: Need to deploy the app to a cloud platform such as AWS.
For scaling: Need to scale the app horizontally or vertically to handle increased traffic and load. 
For Monitoring: Need to  monitor the app for performance and errors using tools such as New Relic.
But now I do not have enough skill to do that.
5) Done
6) Done


Instructions
____________________________________________________
Since the program used MySQL db, need to first install MySQL db for established connection (user@password with promotionsDB). 
The promotions.csv file must be next to the promotions.go file when the program starts to run.
