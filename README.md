#roomalyzer
Small program that retrieves datapoints from roomalyzer API

##How to use
Create 'config.yml' in the same folder as roomalyzer.exe containing valid token and sensor.
    ''
    token: ""
    sensor: ""
    ''
Run '>roomalyzer.exe -o someoutputfile.csv' someoutputfile.csv should now contain the last 48 hours of datapoints.