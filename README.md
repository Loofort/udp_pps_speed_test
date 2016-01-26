# Purpose of test

I wrote this test to compare performance on linux network stack between GO and C. 
In folder `c_send_recv` you can find highly optimized udp sender and receiver written in C, this code was borrowed from this project https://blog.cloudflare.com/how-to-receive-a-million-packets/ and used as beau ideal that GOlang implementation wants to achieve.

Go implements several methods to send UDP packets. There is a test for each - see usage in main.go.
Different servers show different results so you should test your server by yourself.
But in general: 
 * the UDPConn.Write() is the slowest, on my box it almost 2 times slower than C sender.
 * the sendmmsg syscall implementation is fastest and acts the same perfomance as C (even 5% faster on my box)
