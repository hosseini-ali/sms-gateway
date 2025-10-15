vegeta attack -rate=1000 -duration=1s -targets=targets.txt > results.bin
vegeta report -type=text < results.bin
vegeta plot < results.bin > plot.html    # open plot.html in browser
