go run a2as.go --in ../../../../goapple2/source/redbook/monitor.asm --out monitor.rom --flavor redbooka --listing monitor.lst --prefix=-1
go run a2as.go --in ../../../../goapple2/source/redbook/miniasm.asm --out miniasm.rom --flavor redbooka --listing miniasm.lst --prefix=-1
go run a2as.go --in ../../../../goapple2/source/redbook/sweet16.asm --out sweet16.rom --flavor redbooka --listing sweet16.lst --prefix=-1
go run a2as.go --in ../../../../goapple2/source/redbook/fp.asm --out fp.rom --flavor redbookb --listing fp.lst --prefix=-1
go run a2as.go --in ../../../../goapple2/source/redbook/misc-f669.asm --out misc-f669.rom --flavor redbooka --listing misc-f669.lst --prefix=0
go run a2as.go --in ../../../../goapple2/source/redbook/intbasic.asm --out intbasic.rom --flavor merlin --listing intbasic.lst --prefix=-1
go run a2as.go --in ../../../../goapple2/source/applesoft/S.acf --out applesoft.rom --flavor scma --listing applesoft.lst --prefix=-1

MD5_APPLESOFT=$(md5 -q applesoft.rom)
[[ $MD5_APPLESOFT == '84bfbe89c9cd96e589c4d4cb01df4c4a' ]] || echo 'Wrong checksum for applesoft.rom'
MD5_FP=$(md5 -q fp.rom)
[[ $MD5_FP == '76ae6287e5e96471dc95e95eb93ba06d' ]] || echo 'Wrong checksum for fp.rom'
MD5_INTBASIC=$(md5 -q intbasic.rom)
[[ $MD5_INTBASIC == 'c22d8f7ebb54608c8718b66454ca691f' ]] || echo 'Wrong checksum for intbasic.rom'
MD5_MINIASM=$(md5 -q miniasm.rom)
[[ $MD5_MINIASM == 'e64882d56c485ee88d2bfaf4b642c2f9' ]] || echo 'Wrong checksum for miniasm.rom'
MD5_MISC_F669=$(md5 -q misc-f669.rom)
[[ $MD5_MISC_F669 == 'eccaef17e6340b54c309b87ffb6f6f22' ]] || echo 'Wrong checksum for misc-f669.rom'
MD5_MONITOR=$(md5 -q monitor.rom)
[[ $MD5_MONITOR == 'bc0163ca04c463e06f99fb029ad21b1f' ]] || echo 'Wrong checksum for monitor.rom'
MD5_SWEET16=$(md5 -q sweet16.rom)
[[ $MD5_SWEET16 == '93e148f5e30cdd574fd1bb3c26798787' ]] || echo 'Wrong checksum for sweet16.rom'
