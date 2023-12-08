When we add the components we need the following information for each docker component:
Deployment information
    Name of the service.
    Image name of the service.
    Environment mapping
    Port mappings
    Volume mappings
    Secrets list
    Networks list
    Depends list

We can accept this information from:
1. Provide a fields for each information and then accept the value -> This require effort and is limited.
2. Provde a text box which accept the service configuration -> This is good but writing or copying a service configuration could create problem.
3. Provide a button that accept the configuration file -> This file could be invalid or may contain configuration for the additional service, which may we dont want.
Note: After accepting the configuraton we will validate the configuration using the following command.
This means we will first take all the files and then merge them into single file and then use it.
$ docker-compose -f docker-compose.yml config

Note: We need to control the resources, here we dont have the controls.

Once we have configuration for all the components we can deploy them.

Once we have the deployment now we need to control the data from one to another.

The following processing we do with the data.
1. Data acquisition. We acquire the data from different sources.
- The connection to the sources differ.
- In our case we assume its a web services so we acquire data using the api.
- For acquiring data using API we need to call the apis in sequence.
   - Get the token from the authentication service.
   - Use that token to call another api.
   - Use that token to call sequence of api to get the data.
Note: Data may be simple or complex. Its a single value or array of value.
Data will be reproduce on timely base or whenever required.

2. Once the data is obtained. There are some validation is required on that data.
Not all data are required. So we filter the value we need.
Sometimes data are split or merged based on the requirement.
Sometimes data are translated into another component.

3. Sometime we need to acquire the data from multiple sources. 
In this case we need to wait till we have the data from all the source.
Now the data production for each resource will have different frequency. 
Here we need to define whether we need unique data or duplicate data will do.
Secondly, we need to validate the time. 

4. After acquiring the data we need to pass it to processing service. 
Here the service may accept the data in a particular format.
So the conversion from the source to the detination format must be done by orchestrator.

