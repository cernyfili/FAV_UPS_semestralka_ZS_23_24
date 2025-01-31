# Klient

## Dekompozice do modulů

- **main** - hlavní modul, který spouští klienta

### Backend

- **parser** - parsování vstupních dat, přímo převod stringu na interní reprezentaci dat a naopak
- **server_communication** - komunikace se serverem, posílání a přijímání dat
- **state_manager** - implementuje stavový automat, který řídí chování klienta

### Frontend

- **views**
    - **start_page** - úvodní stránka, kde se zadává adresa serveru a jméno hráče
    - **lobby** - stránka, kde se čeká na ostatní hráče a lze vytvořit nebo připojit se do hry
    - **before_game** - stránka, kde se čeká na začátek hry nebo může být hra zahájena
    - **running_game** - stránka, kde se hraje hra
    - **my_turn_roll_dice** - kolo hráče, kde hází kostkou
    - **my_turn_select_cubes** - kolo hráče, kde vybírá kostky, které chce ponechat
    - následně _helper funkce_, které se používají v různých views a _interface_ pro tyta views
- **ui_manager** - řídí zobrazení views a komunikaci mezi nimi

### Shared

- konstanty a funkce, které jsou používány jak v backendu, tak ve frontendu

# Rozvrstvení aplikace

- **server_communication** - je hlavní aplikace, která používá stavový automat a parser a reaguje na příchozí zprávy

# Použité technologie

## Knihovny

- **Tkinter** - pro _gui_ klienta
- **stateless** - pro stavový automat

## Další

- Python 3.10 a standardní knihovny

# Metody paralelizace

- gui
    - využívá více vláken aby současně mohlo přijímat zprávy a zobrazovat data
    - opakovaně poslouchá příjem zpráv a následně zobrazuje data
