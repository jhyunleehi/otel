 fio --name=mytest --ioengine=sync --rw=randrw --bs=4k --size=100M --numjobs=3 --time_based --runtime=300m --output=fio_random_rw.txt