# simlife
"The Life Game" simulation engine written in golang. Originally it was the task for a job application interview. Later it was extended with pseudo-graphic console UI, several implementations of the main calculation algorithm with different optimizations, including a multithreaded implementation and comparative performance tests.

This project is solely educational, with the purpose of demonstrating certain aspects of the Go programming language like, how to:
 - Create a console application with pseudo-graphic console UI with windows and resizable layouts
 - Use keyboard and mouse in the console terminal
 - Process startup parameters including commands and flags
 - Use goroutines to utilize all CPU cores efficiently by performing calculations simultaneously in several threads
 - Use channels to synchronize goroutines

![vokoscreen-2021-09-15_22-14-48](https://user-images.githubusercontent.com/41936843/133479356-913399ff-181c-4b74-9a67-8bd9d4e5755a.gif)

Console UI
![Screenshot from 2021-09-15 20-13-14](https://user-images.githubusercontent.com/41936843/133461802-475c712e-6fb3-4b7b-b560-e9ee344a19bc.png)

Startup Flags
![Screenshot from 2021-09-16 16-40-52](https://user-images.githubusercontent.com/41936843/133605908-f2d4b339-ffad-4d92-a8c7-79a60c1065e0.png)

Simulation with short console output
![Screenshot from 2021-09-16 16-47-35](https://user-images.githubusercontent.com/41936843/133606983-cdf62078-286c-4074-ae22-1aa4c6e0981d.png)

Simulation with multithreading mode with 10 workers (2.5 times faster compared with standard mode)
![Screenshot from 2021-09-16 16-46-25](https://user-images.githubusercontent.com/41936843/133607067-78a65986-16eb-428e-a236-d58d47556926.png)

Due to lack of available time to this project, there are a few unsightly, poorly formatted and potentially problematic pieces in this code. It is strongly advised not to use this code in a production environment :-)
