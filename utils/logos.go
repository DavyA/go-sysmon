package utils

import "strings"

type Logo struct {
	Art   string
	Color string
}

func GetOSLogo(osName string, kernel string) Logo {
	lowerOS := strings.ToLower(osName)
	lowerKrn := strings.ToLower(kernel)

	// WSL Detection
	if strings.Contains(lowerKrn, "microsoft") || strings.Contains(lowerKrn, "wsl") {
		return Logo{
			Color: "\033[36m", // Cyan
			Art: `
                                ..,
                    ....,,:;+ccllll
      ...,,+:;  cllllllllllllllllll
,cclllllllllll  lllllllllllllllllll
llllllllllllll  lllllllllllllllllll
llllllllllllll  lllllllllllllllllll
llllllllllllll  lllllllllllllllllll
llllllllllllll  lllllllllllllllllll
llllllllllllll  lllllllllllllllllll
                                   
llllllllllllll  lllllllllllllllllll
llllllllllllll  lllllllllllllllllll
llllllllllllll  lllllllllllllllllll
llllllllllllll  lllllllllllllllllll
llllllllllllll  lllllllllllllllllll
` + "`" + `'ccllllllllll  lllllllllllllllllll
      ` + "`" + `'*::  :ccllllllllllllllll
                       ` + "`" + `''*::cll
                                 `,
		}
	}

	if strings.Contains(lowerOS, "ubuntu") {
		return Logo{
			Color: "\033[38;5;208m", // Orange
			Art: `
            .-/+oossssoo+/-.
        ` + "`" + `:+ssssssssssssssssss+:` + "`" + `
      -+ssssssssssssssssssyyssss+-
    .ossssssssssssssssssdMMMNysssso.
   /ssssssssssshdmmNNmmyNMMMMhssssss/
  +ssssssssshmydMMMMMMMNddddyssssssss+
 /sssssssshNMMMyhhyyyyhmNMMMNhssssssss/
.ssssssssdMMMNhsssssssssshNMMMdssssssss.
+sssshhhyNMMNyssssssssssssyNMMMysssssss+
ossyNMMMNyMMhsssssssssssssshmmmhssssssso
ossyNMMMNyMMhsssssssssssssshmmmhssssssso
+sssshhhyNMMNyssssssssssssyNMMMysssssss+
.ssssssssdMMMNhsssssssssshNMMMdssssssss.
 /sssssssshNMMMyhhyyyyhdNMMMNhssssssss/
  +sssssssssdmydMMMMMMMMddddyssssssss+
   /ssssssssssshdmNNNNmyNMMMMhssssss/
    .ossssssssssssssssssdMMMNysssso.
      -+sssssssssssssssssyyyssss+-
        ` + "`" + `:+ssssssssssssssssss+:` + "`" + `
            .-/+oossssoo+/-.`,
		}
	} else if strings.Contains(lowerOS, "arch") || strings.Contains(lowerOS, "manjaro") {
		return Logo{
			Color: "\033[36m", // Cyan
			Art: `
                   -` + "`" + `
                  .o+` + "`" + `
                 ` + "`" + `ooo/
                ` + "`" + `+oooo:
               ` + "`" + `ooooooo.
              ` + "`" + `+ooooooo+` + "`" + `
              +ooooooooo` + "`" + `
             +ooooooooooo` + "`" + `
            .ooooooooooooo` + "`" + `
           ` + "`" + `ooooooooooooooo` + "`" + `
          ` + "`" + `ooooooooooooooooo` + "`" + `
         .ooooooooooooooooooo.
        .ooooooooooooooooooooo.
       .ooooooooooooooooooooooo.
      .ooooooooooooooooooooooooo.
     .ooooooooooooooooooooooooooo.
    .ooooooooooooooooooooooooooooo.
   .ooooooooooooooooooooooooooooooo.
  ` + "`" + `ooooooooooooooooooooooooooooooooo` + "`" + `
 ` + "`" + `ooooooooooooooooooooooooooooooooooo` + "`" + ``,
		}
	} else if strings.Contains(lowerOS, "fedora") {
		return Logo{
			Color: "\033[34m", // Blue
			Art: `
          /:-------------:\
       :-------------------::
     :-----------/shhOHbmp---:\
   /-----------omMMMNNNMMD  ---:
  :-----------sMMMMNMNMP.    ---:
 /-----------:MMMdP-------    ---\
,------------:MMMd--------    ---:
:------------:MMMd-------    .---:
:----    oNMMMMMMMMMNho     .----:
:--     .+shhhMMMmhhy++   .------/
:-    -------:MMMd--------------:
:-   --------/MMMd-------------;
:-    ------/hMMMy------------:
:-- :dMNdhhdNMMNo------------;
:---:sdNMMMMNds:------------:
:------:://:-------------::
 :---------------------://`,
		}
	} else if strings.Contains(lowerOS, "debian") {
		return Logo{
			Color: "\033[31m", // Red
			Art: `
       _,met$$$$$gg.
    ,g$$$$$$$$$$$$$$P.
  ,g$$P"     """Y$$.".
 ,$$P'              ` + "`" + `$$$.
',$$P       ,ggs.     ` + "`" + `$$b:
` + "`" + `d$$'     ,$P"'   .    $$$
 $$P      d$'     ,    $$P
 $$:      $$.   -    ,d$$'
 $$;      Y$b._   _,d$P'
 Y$$.    ` + "`" + `. ` + "`" + `"Y$$$$P"'
 ` + "`" + `$$b      "-.__
  ` + "`" + `Y$$
   ` + "`" + `Y$$.
     ` + "`" + `$$b.
       ` + "`" + `Y$$b.
          ` + "`" + `"Y$b._
              ` + "`" + `"""`,
		}
	} else if strings.Contains(lowerOS, "darwin") || strings.Contains(lowerOS, "macos") || strings.Contains(lowerOS, "mac") {
		return Logo{
			Color: "\033[37m", // White
			Art: `
                    'c.
                 ,xNMM.
               .OMMMMo
               OMMM0,
     .;loddo:' loolloddol;.
   cKMMMMMMMMMMNWMMMMMMMMMM0:
 .KMMMMMMMMMMMMMMMMMMMMMMMWd.
 XMMMMMMMMMMMMMMMMMMMMMMMX.
;MMMMMMMMMMMMMMMMMMMMMMMM:
:MMMMMMMMMMMMMMMMMMMMMMMM:
.MMMMMMMMMMMMMMMMMMMMMMMMX.
 kMMMMMMMMMMMMMMMMMMMMMMMMWd.
 .XMMMMMMMMMMMMMMMMMMMMMMMMMMk
  .XMMMMMMMMMMMMMMMMMMMMMMMMK.
    kMMMMMMMMMMMMMMMMMMMMMMd
     ;KMMMMMMMWXXWMMMMMMMk.
       .cooc,.    .,coo:.`,
		}
	} else if strings.Contains(lowerOS, "suse") {
		return Logo{
			Color: "\033[32m", // Green
			Art: `
           .;ldkO0000Okdl;.
       .;d00xl:^'''''^:lx00d;.
     .d00l'                'l00d.
   .d0Kl'                    'lK0d.
  .k0l.                        .l0k.
 .k0l.                          .l0k.
 '0Kl.                          .lK0'
 '0Kx.          'lxxxxx.        .xK0'
 '0K0o.        'xK00000.       .o0K0'
  .x00l.      'xK00000'       .l00x.
   .xK0o'     'xK0000.      'o0Kx.
     .o00x:'          ':x00o.
       .lx00kdl::c::ldk00xl.
           .;ldkO0000Okdl;.`,
		}
	} else if strings.Contains(lowerOS, "mint") {
		return Logo{
			Color: "\033[32m", // Green
			Art: `
             ...-:::::-...
          .-MMMMMMMMMMMMMMM-.
      .-MMMM` + "`" + `..-:::::::-..` + "`" + `MMMM-.
    .:MMMM.:MMMMMMMMMMMMMMM:.MMMM:.
   -MMM-M---MMMMMMMMMMMMMMMMMMM.MMM-
 ` + "`" + `MMMM:MM` + "`" + `  -MMMMMMMMMMMMMMMMMMMM:MMMM` + "`" + `
.MMM.NMMMM+  ` + "`" + `..---::---..` + "`" + ` ` + "`" + `::-. .MMM.
:MMN-MMMMMM   MMMMMMMMMMMMM- -MMN-MMN:
:MMM-MMMMMM   MMMMMMMMMMMMM. .MMM-MMM:
:MMM-MMMMMM   MMMMMMMMMMMMM. .MMM-MMM:
:MMM-MMMMMM   MMMMMMMMMMMMM. .MMM-MMM:
:MMM-MMMMMM   MMMMMMMMMMMMM. .MMM-MMM:
` + "`" + `MMM:MMM:MMMM   MMMMMMMMMMMMM. MMM:MMM` + "`" + `
 -MMM-M- .MMM-MMMMMMMMMMMMM. MMM-MMM-
  .:MMMM:. ` + "`" + `MMMMMMMMMMMMM-MMMMMM:.
    .-MMMM` + "`" + `..--::::::--..` + "`" + `MMMM-.
       .-MMMMMMMMMMMMMMMMM-.
          ` + "`" + `...-:::::-...` + "`" + ``,
		}
	} else if strings.Contains(lowerOS, "alpine") {
		return Logo{
			Color: "\033[34m", // Blue
			Art: `
       .hddddddddddddddddddddddh.
      :dddddddddddddddddddddddddd:
     /dddddddddddddddddddddddddddd/
    +dddddddddddddddddddddddddddddd+
  ` + "`" + `sdddddddddddddddddddddddddddddddds` + "`" + `
 ` + "`" + `ydddddddddddd++hdddddddddddddddddddy` + "`" + `
.mddddddddddd+  .mdddddddddddddddddddm.
/dddddddddd+    .mdddddddddddddddddddd/
+ddddddddd+     .mdddddddddddddddddddd+
+dddddddd+      .mdddddddddddddddddddd+
/dddddddddd+    .mdddddddddddddddddddd/
.mddddddddddd+  .mdddddddddddddddddddm.
 ` + "`" + `ydddddddddddd++hdddddddddddddddddddy` + "`" + `
  ` + "`" + `sdddddddddddddddddddddddddddddddds` + "`" + `
    +dddddddddddddddddddddddddddddd+
     /dddddddddddddddddddddddddddd/
      :dddddddddddddddddddddddddd:
       .hddddddddddddddddddddddh.`,
		}
	} else if strings.Contains(lowerOS, "gentoo") {
		return Logo{
			Color: "\033[35m", // Purple
			Art: `
         -/oyddmdhs+:.
     -oNMMMMMMMMNNmhy+-
   -yNMMMMMMMMMMMNNNdy.
 ` + "`" + `oNMMMMMMMMMMMMNNNmho.
.dMMMMMMMMMMMMMMNNNme.
-mMMMMMMMMMMMMMMNNNN+
-NMMMMMMo..:/+ossyyy+
.dMMMMMMh     ` + "`" + `.-://.
 ` + "`" + `+mMMMMMMNmmNNNMh-
   -smMMMMMMMMNh-
      -sdmNNmds-`,
		}
	} else if strings.Contains(lowerOS, "redhat") || strings.Contains(lowerOS, "rhel") || strings.Contains(lowerOS, "centos") {
		return Logo{
			Color: "\033[31m", // Red
			Art: `
           .MMM..:MMMMMMM
          MMMMMMMMMMMMMMMMMM
          MMMMMMMMMMMMMMMMMMMM.
         MMMMMMMMMMMMMMMMMMMMMM
        ,MMMMMMMMMMMMMMMMMMMMMM:
        MMMMMMMMMMMMMMMMMMMMMMMM
  .MMMM'  MMMMMMMMMMMMMMMMMMMMMM
 MMMMMM    ` + "`" + `MMMMMMMMMMMMMMMMMMMM.
MMMMMMMM      MMMMMMMMMMMMMMMMMM .
MMMMMMMMM.       ` + "`" + `MMMMMMMMMMMMM' MMM.
MMMMMMMMMMM.                     MMMM
` + "`" + `MMMMMMMMMMMMM.                 ,MMMMM.
 ` + "`" + `MMMMMMMMMMMMMMMMM.          ,MMMMMMMM.
    MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM.
      MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM.
        ` + "`" + `MMMMMMMMMMMMMMMMMMMMMMMMMMMMMM.
           ` + "`" + `MMMMMMMMMMMMMMMMMMMMMMMMMM.
              ` + "`" + `MMMMMMMMMMMMMMMMMMMMMM.
                  ` + "`" + `MMMMMMMMMMMMMMMMM.`,
		}
	}

	// Default Linux (Penguin / stylized)
	return Logo{
		Color: "\033[36m", // Cyan
		Art: `
        :::
      :::::     
    :::  :::    
    :::  :::    
    :::::::     
    :::  :::    
    :::  :::    
    :::  :::    
                
 :::     :::  
  :::   :::   
   :::::::    
    ::::      
    ::::      
   :::::::    
  :::   :::   
 :::     :::  `,
	}
}
