# Ouroboros Release

Ouroboros release is a Bosh release for putting heavy load on Loggregator.

The name Ouroboros comes from the [serpent eating itself](ouroboros).  With
such a name, the ouroboros job reads from the firehose of one Loggregator
deployment and writes all the log traffic out into another Loggregator
deployment. The ouroboros process is simply a firehose nozzle.

In addition to ouroboros, there is volley, a hostile consumer of Loggregator.
The purpose of volley is to ensure Loggregator can withstand bad actors.

Finally, there is syslogr which provides a variety of poorly performing syslog
drains. The syslogr process adds further load on Loggregator.

[ouroboros]: https://en.wikipedia.org/wiki/Ouroboros
