# Welcome to the proof of concept weather app in Go!

### to run
1. Git pull the repo, navigate to the project and run `EMAIL={contactEmail} go run main.go
2. In a browser or curl, run `http://localhost:8080/forecast/41.8832/-87.6324` or any other coordinates (these ones are for Chicago)
3. Verify the correctness of the weather data by checking the weather for those coordinates via a trusted source!

### improvements
This is intended as a proof of concept only, and as such lacks some features.
In order, I would add
1. caching, potentially with rounding of the coordinate data to avoid spamming weather.gov
2. either validation that the coordinates are in a supported country or logic to get weather data from other countries
3. More weather categories! I originally had chilly, Chicago chilly, and TOO HOT as temp ranges but I reread the assignment and noticed that there were 3 specifically listed categories. 
4. Logic to deduce how much a given level of wind alters temperature perception, and use that to modify the temperature ranges accordingly. 

Eventually I would profile the code to make a determination of what steps would increase throughput, although as it pretty much just acts as a proxy I don't expect to find many problems. 
