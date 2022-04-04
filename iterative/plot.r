dataParallel=read.csv('/tmp/result-parallel.csv')
dataSerial=read.csv('/tmp/result-serial.csv')

percentile <- seq(0.1, 0.99, by = 0.01)
q <- quantile(x = dataParallel$y, probs = percentile)

plot(x = percentile,y = q,
     axes = FALSE,
	 xlim=c(0.1, 1),
	 ylim=c(0, 4),
     xlab = "percentile",
     ylab = "latency",
     main = "Plot",
     col="red",
     type="l"
)
axis(side = 1, at = seq(0, 0.99, by = 0.01))
axis(side = 2, at = seq(0, 4, by = 0.1))
box()

v <- quantile(x = dataSerial$y, probs = percentile)

lines(x = percentile,y = v,
     main = "Plot",
     col="green"
)
