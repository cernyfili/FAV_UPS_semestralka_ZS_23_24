# Server
## Dekompozice do modulů
- **internal/main.go** - je vsupním bodem aplikace, zde se nachází funkce `main` a zde se volají všechny ostatní funkce.
- **Command Processing** - zpracování příkazů od uživatele
- **Models** - datové struktury potřebné pro běh aplikace
  - _State Machine_ - stavový automat, který reprezentuje stav hry
  - _Player_- uživatel aplikace a připojení
  - _Game_- hra a její stav
  - _Message_- zpráva, kterou klient posílá serveru
  - a jejich listy pro uchování více instancí
- **Network** - síťová komunikace - odesílání a příjem zpráv
- **Parser** - zpracování zpráv od serveru a následně vnitřní objekty na posílané zprávy
- **Server Listen** - hlavní smyčka, která naslouchá zprávám od serveru a zpracovává je pomocí funkcí z jiných modulů
# Použité technologie
## Knihovny
- **stateless** - knihovna pro vytváření stavových automatů
- **github.com/pkg/errors** - knihovna pro zachycení chyb
- **github.com/sirupsen/logrus** - knihovna pro logování
## Další
- Go 1.23
# Metody paralelizace
- **Komunikace s klientem** - pro každého klienta je vytvořen samostatný vlákno, které naslouchá zprávám od klienta a také posílá zprávy klientovi
- **Ukončení spojení** - před naprostým odpojením klienta má klient možnost připojit se zpět do hry a to díky pomocí vlákna, které po uplynutí určité doby za podmínky, že se klient nepřipojí, ukončí spojení

- jsou využívány gorutiny pro asynchronní zpracování zpráv od serveru

# Spuštění
## Server
1. Nainstalujte Go 1.23, make
2. Stáhněte si zdrojový kód
3. Přejděte do složky `GameServer`
4. Spusťte aplikaci pomocí příkazu `make`

## Klient
1. Nainstalujte Python 3.1, make
2. Stáhněte si zdrojový kód
3. Přejděte do složky `GameClient`
4. Spusťte aplikaci pomocí příkazu `make`

