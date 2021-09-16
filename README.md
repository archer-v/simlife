# simlife
"The Life Game" simulation engine written in golang. Originally it was the test task for a job application interview. Later it was extented "for fun" with additional features: 
 - command line and interactive mode
 - console UI with responsive panel and layouts, keyboard and mouse support
 - confurable startup parameters by command-line flags
 - several implementations of the main calculation algorithm, including multithreaded implementation
 - benchmark tests

![vokoscreen-2021-09-15_22-14-48](https://user-images.githubusercontent.com/41936843/133479356-913399ff-181c-4b74-9a67-8bd9d4e5755a.gif)

I don't think that this project can be useful for something except learning. I hope it will useful if you looking for a golang examples of how to:
 - Write console application with UI, pseudo-graphics windows, and resizable layouts
 - Use keyboard and mouse in the console terminal
 - Process startup parameters including commands and flags
 - Use goroutines to utilize all CPU cores efficiently by performing calculations simultaneously in several threads
 - Use channels to synchronize goroutines

Console UI
![Screenshot from 2021-09-15 20-13-14](https://user-images.githubusercontent.com/41936843/133461802-475c712e-6fb3-4b7b-b560-e9ee344a19bc.png)

Startup Flags
![Screenshot from 2021-09-16 16-40-52](https://user-images.githubusercontent.com/41936843/133605908-f2d4b339-ffad-4d92-a8c7-79a60c1065e0.png)

Simulation with short console output
![Screenshot from 2021-09-16 16-47-35](https://user-images.githubusercontent.com/41936843/133606983-cdf62078-286c-4074-ae22-1aa4c6e0981d.png)

Simulation with multithreading mode with 10 workers (2.5 times faster compared with standard mode)
![Screenshot from 2021-09-16 16-46-25](https://user-images.githubusercontent.com/41936843/133607067-78a65986-16eb-428e-a236-d58d47556926.png)

I have no free time to do comprehensive code decorating according to all modern practices. Also, there are a few ugly and potentially problematic peaces in this code. Do not use this code in production :-)
