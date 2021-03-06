go build a2as.go

echo monitor.rom
./a2as --in ../../../../goapple2/source/redbook/monitor.asm --out monitor.rom --flavor redbooka --listing monitor.lst --prefix=-1
MD5_MONITOR=$(md5 -q monitor.rom)
[[ $MD5_MONITOR == 'bc0163ca04c463e06f99fb029ad21b1f' ]] || (echo 'Wrong checksum for monitor.rom'; false) || exit 1
rm -f monitor.rom monitor.lst

echo autostart.rom
./a2as --in ../../../../goapple2/source/redbook/autostart.asm --out autostart.rom --flavor redbooka --listing autostart.lst --prefix=-1
MD5_AUTOSTART=$(md5 -q autostart.rom)
[[ $MD5_AUTOSTART == '8925b695ae0177dd3919dbea2f2f202b' ]] || (echo 'Wrong checksum for autostart.rom'; false) || exit 1
rm -f autostart.rom autostart.lst

echo miniasm.rom
./a2as --in ../../../../goapple2/source/redbook/miniasm.asm --out miniasm.rom --flavor redbooka --listing miniasm.lst --prefix=-1
MD5_MINIASM=$(md5 -q miniasm.rom)
[[ $MD5_MINIASM == 'e64882d56c485ee88d2bfaf4b642c2f9' ]] || (echo 'Wrong checksum for miniasm.rom'; false) || exit 1
rm -f miniasm.rom miniasm.lst

echo sweet16.rom
./a2as --in ../../../../goapple2/source/redbook/sweet16.asm --out sweet16.rom --flavor redbooka --listing sweet16.lst --prefix=-1
MD5_SWEET16=$(md5 -q sweet16.rom)
[[ $MD5_SWEET16 == '93e148f5e30cdd574fd1bb3c26798787' ]] || (echo 'Wrong checksum for sweet16.rom'; false) || exit 1
rm -f sweet16.rom sweet16.lst

echo fp.rom
./a2as --in ../../../../goapple2/source/redbook/fp.asm --out fp.rom --flavor redbookb --listing fp.lst --prefix=-1
MD5_FP=$(md5 -q fp.rom)
[[ $MD5_FP == '76ae6287e5e96471dc95e95eb93ba06d' ]] || (echo 'Wrong checksum for fp.rom'; false) || exit 1
rm -f fp.rom fp.lst

echo misc-f669.rom
./a2as --in ../../../../goapple2/source/redbook/misc-f669.asm --out misc-f669.rom --flavor redbooka --listing misc-f669.lst --prefix=0
MD5_MISC_F669=$(md5 -q misc-f669.rom)
[[ $MD5_MISC_F669 == 'eccaef17e6340b54c309b87ffb6f6f22' ]] || (echo 'Wrong checksum for misc-f669.rom'; false) || exit 1
rm -f misc-f669.rom misc-f669.lst

echo intbasic.rom
./a2as --in ../../../../goapple2/source/redbook/intbasic.asm --out intbasic.rom --flavor merlin --listing intbasic.lst --prefix=-1
MD5_INTBASIC=$(md5 -q intbasic.rom)
[[ $MD5_INTBASIC == 'c22d8f7ebb54608c8718b66454ca691f' ]] || (echo 'Wrong checksum for intbasic.rom'; false) || exit 1
rm -f intbasic.rom intbasic.lst

echo applesoft.rom
./a2as --in ../../../../goapple2/source/applesoft/S.acf --out applesoft.rom --flavor scma --listing applesoft.lst --prefix=-1
MD5_APPLESOFT=$(md5 -q applesoft.rom)
[[ $MD5_APPLESOFT == '84bfbe89c9cd96e589c4d4cb01df4c4a' ]] || (echo 'Wrong checksum for applesoft.rom'; false) || exit 1
rm -f applesoft.rom applesoft.lst

echo hires.rom
./a2as --in ../../../../goapple2/source/progaid/hires.asm --out hires.rom --flavor redbooka
MD5_HIRES=$(md5 -q hires.rom)
[[ $MD5_HIRES == 'efe22f1a8c94458068fb12ae702b58c4' ]] || (echo 'Wrong checksum for hires.rom'; false) || exit 1
rm -f hires.rom

echo verify.rom
./a2as --in ../../../../goapple2/source/progaid/verify.asm --out verify.rom --flavor redbooka
MD5_VERIFY=$(md5 -q verify.rom)
[[ $MD5_VERIFY == '527f420462426e4851b942af46cc7f48' ]] || (echo 'Wrong checksum for verify.rom'; false) || exit 1
rm -f verify.rom

echo ramtest.rom
./a2as --in ../../../../goapple2/source/progaid/ramtest.asm --out ramtest.rom --flavor redbooka
MD5_RAMTEST=$(md5 -q ramtest.rom)
[[ $MD5_RAMTEST == '0420635256a3b016323989e3a9fe4ce7' ]] || (echo 'Wrong checksum for ramtest.rom'; false) || exit 1
rm -f ramtest.rom

echo music.rom
./a2as --in ../../../../goapple2/source/progaid/music.asm --out music.rom --flavor redbooka
MD5_MUSIC=$(md5 -q music.rom)
[[ $MD5_MUSIC == '0ffe796a73410e822fcae5e510374924' ]] || (echo 'Wrong checksum for music.rom'; false) || exit 1
rm -f music.rom
