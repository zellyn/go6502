go run a2as.go --in ../../../../goapple2/source/redbook/monitor.asm --out monitor.rom --flavor redbooka --listing monitor.lst --prefix=-1
go run a2as.go --in ../../../../goapple2/source/redbook/miniasm.asm --out miniasm.rom --flavor redbooka --listing miniasm.lst --prefix=-1
go run a2as.go --in ../../../../goapple2/source/redbook/sweet16.asm --out sweet16.rom --flavor redbooka --listing sweet16.lst --prefix=-1
go run a2as.go --in ../../../../goapple2/source/redbook/fp.asm --out fp.rom --flavor redbookb --listing fp.lst --prefix=-1
go run a2as.go --in ../../../../goapple2/source/redbook/misc-f699.asm --out misc-f699.rom --flavor redbooka --listing misc-f699.lst --prefix=0
go run a2as.go --in ../../../../goapple2/source/redbook/intbasic.asm --out intbasic.rom --flavor merlin --listing intbasic.lst --prefix=-1

