cat writes.txt |
grep Wrote |
sed -e 's/:.*//' |
grep -ve '1094\|01..\|005[DE]\|023[34567]\|000F\|001[0123]\|3513\|3534' |
head -10
