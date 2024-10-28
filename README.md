## Nombres y rol:
+ Sebastián Arrieta, 202173511-9
+ Jonathan Olivares, 202073096-2

## Sobre el uso:
- Los nodos regionales comenzarán a enviar datos automaticamente al inciarse.
- Solo el nodo tai requiere que se le entregue una opción por consola.
- La distribución de los servicios es la siguiente:
  + En la maquina virtual 1 (MV1): Primary Node y DataNode1
  + En la maquina virtual 2 (MV2): DataNode2 y Isla-File
  + En la maquina virtual 3 (MV3): Continente Folder y Nodo Tai
  + En la maquina virtual 4 (MV4): Continente Server y Diaboromon
  
## Para la ejecución
- Los docker utilizan docker compose up
- Estos docker son ejecutados mediante un makefile.
- Para la ejecución del makefile se puede usar el comando `make run` en cada maquina virtual.
- **El orden de ejecución de las maquinas virtuales es:  MV1 -> MV2 -> MV3 -> MV4.**
- Antes de comenzar a tener interacción con el programa se debe esperar que estén todos los servicios corriendo de acuerdo al comando `make run`.
- MV3 posee al Nodo Tai, al necesitar un input usando docker compose se hace un attach a ese servicio especifico por lo que no se muestra el log del otro servicio. 
- Al hacer attach al Nodo Tai no se muestra el primer print, de recomendación enviar el numero 1 como input para que aparezca nuevamente le menu, pero de igual forma es un menu que dice lo siguiente: 
  - Seleccione una opción:
  - 1 Pedir datos sacrificados
  - 2 Atacar a Diaboromon
  - 0 Salir

- El programa finaliza cuando diaboromon o nodo tai son derrotados. Tambien finaliza cuando se le da a la opcion de salir en el menu del nodo tai, en este caso se asume que Tai pierde. Cuando sucede una de las condiciones de finalización se manda un mensaje para que los demás servicios vayan terminando.

- **Para lograr ejecutar los servicios se debe usar sudo al usar los makefile**
