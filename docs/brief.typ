#set document(
  title: "Milpa: Eco-Mercado Digital",
  author: "Equipo Hackathon Nicaragua 2026",
)

#let proyecto = "Milpa"
#let version = "2.0"

= Milpa: Eco-Mercado Digital
Especificación de Requisitos de Software

Proyecto: #proyecto \
Versión: #version \
Fecha: Julio 2026

#pagebreak()

= Ficha del documento

#table(
  columns: 4,
  table.header("Fecha", "Revisión", "Autor", "Descripción"),
  [5/Jul/26], [1.0], [Oscar Reyes], [Propuesta inicial (reto incorrecto)],
  [26/Jul/26], [2.0], [Oscar Reyes], ["Reescritura completa para reto Eco-Mercado Digital"],
)

#pagebreak()

#outline()

#pagebreak()

= 1. Introducción

== 1.1 Propósito

Este documento define los requisitos funcionales y no funcionales de _Milpa_, una plataforma digital de mercado agropecuario diseñada para conectar directamente a pequeños y medianos agro-productores de frutas y cítricos con compradores minoristas y mayoristas en Nicaragua.

El sistema busca eliminar la cadena de intermediación que encarece los productos y reduce el beneficio del productor, permitiendo un canal directo de venta que disminuya el desperdicio de cosechas y ofrezca precios más justos para ambas partes.

La audiencia de este documento incluye:

- Equipo de desarrollo (arquitectura e implementación del servidor).
- Mentores y jurado evaluador del Hackathon Nicaragua 2026.
- Diseñadores UX/UI (aplicación web y móvil).
- Stakeholders del proyecto.
- Futuros mantenedores del sistema.

== 1.2 Alcance

_Milpa_ es una plataforma digital cuyo propósito es funcionar como un mercado digital para la comercialización directa de productos frutales y cítricos entre agricultores y compradores.

=== El sistema permitirá:

- Registro y autenticación de usuarios (agricultores y compradores).
- Gestión de perfiles de agricultor, incluyendo registro de empresas agrícolas con verificación.
- Creación y administración de catálogos de productos por parte del agricultor.
- Navegación y búsqueda del catálogo por parte de compradores minoristas.
- Chat directo entre comprador minorista y agricultor (sin necesidad de match previo).
- Publicación de solicitudes de abastecimiento por compradores mayoristas.
- Recepción de ofertas de agricultores a solicitudes de abastecimiento.
- Sistema de asignación tipo _match_ (similar a Tinder) basado en cercanía geográfica, precio, tiempo de entrega y reputación.
- Chat habilitado entre mayorista y agricultor posterior a la formación de un match.
- Publicación de ofertas de liquidación de lotes por parte del agricultor (con opción de restringir a compradores mayoristas).
- Sistema de calificaciones mutuas (1 a 5 estrellas) posterior a cada transacción.
- Geolocalización de agricultores y productos.
- Moderación de productos por parte del administrador.
- Sistema de reportes de publicaciones fraudulentas o falsas.

=== El sistema no contempla:

- Procesamiento de pagos electrónicos (las partes acuerdan el método de pago directamente).
- Logística, delivery o gestión de envíos (el agricultor indica si realiza entregas o el producto se retira en punto específico).
- Facturación electrónica o generación de comprobantes fiscales.
- Contratos legales entre las partes.
- Gestión de inventario en tiempo real.

== 1.3 Personal involucrado

#table(
  columns: 3,
  table.header("Nombre", "Rol", "Responsabilidades"),
  [Emily Cáceres], [Comunicación], [Pitch y presentaciones],
  [Jean Cruz], [Diseño UX/UI], [Experiencia de usuario, interfaz web y móvil],
  [Penélope Martínez], [Marketing], [Estrategia de mercado y difusión],
  [Oscar Reyes], [Arquitecto y Desarrollador], [Arquitectura del sistema y API],
  [Misael Vanegas], [Desarrollador], [Desarrollo de clientes de la API],
)

== 1.4 Definiciones, acrónimos y abreviaturas

#table(
  columns: 2,
  table.header("Término", "Definición"),
  ["Agricultor"], ["Productor agrícola que ofrece sus cultivos a través de la plataforma. Puede ser persona individual o empresa formalizada."],
  ["Comprador Minorista"], ["Consumidor final que adquiere productos en pequeñas cantidades para consumo propio."],
  ["Comprador Mayorista"], ["Entidad que adquiere productos en volumen para reventa o procesamiento. Se subdivide en Detallista (mercados) y Corporativo (supermercados, cadenas)."],
  ["Solicitud de Abastecimiento"], ["Publicación creada por un comprador mayorista indicando su necesidad de un producto, cantidad, unidad de medida y fecha límite."],
  ["Oferta de Abastecimiento"], ["Respuesta de un agricultor a una solicitud de abastecimiento, indicando cantidad disponible, precio y tiempo de entrega."],
  ["Oferta de Liquidación"], ["Publicación del agricultor ofreciendo un lote completo de producto a un precio preferencial, usualmente para evitar pérdida por excedente de cosecha."],
  ["Match"], ["Vínculo formal entre un comprador mayorista y un agricultor cuando el comprador acepta una oferta de abastecimiento."],
  ["API"], ["Application Programming Interface — interfaz de programación de aplicaciones."],
  ["JWT"], ["JSON Web Token — estándar para autenticación basada en tokens."],
  ["ERS"], ["Especificación de Requisitos del Sistema."],
  ["MVP"], ["Producto Mínimo Viable."],
  ["RUC"], ["Registro Único de Contribuyente ( Nicaragua)."],
  ["NIT"], ["Número de Identificación Tributaria."],
)

== 1.5 Referencias

- IEEE Std. 830-1998 — _IEEE Recommended Practice for Software Requirements Specifications_.
- Bases del Hackathon Nicaragua 2026.
- Documento de retos del Hackathon Nicaragua 2026 — Reto: Eco-Mercado Digital (Categoría Aficionado, Temática Agropecuario/Medio Ambiente).
- _Eco-Mercado Digital para Agro-Productores en Nicaragua: Situación de los Pequeños Productores de Frutas y Cítricos en Nicaragua_ — Documento de investigación del proyecto.
- Plan Nacional de Lucha contra la Pobreza y para el Desarrollo Humano 2022-2026 (PNCL-DH).

#pagebreak()

= 2. Descripción General

== 2.1 Perspectiva del producto

El sistema es una plataforma compuesta por:

- *API REST* (backend en Go) que expone los servicios del dominio.
- *WebSockets* para el módulo de chat.
- *Aplicación web* (frontend React).
- *Aplicación móvil* (Flutter).
- *Base de datos* PostgreSQL.
- *Capa de caché* Redis.

#figure(
```text
Web ──┐
      ├── API REST ──── PostgreSQL
Móvil─┘        │
               ├── Redis (caché)
               └── WebSocket (chat)
```
)

La plataforma se desarrolla bajo una *arquitectura hexagonal* (puertos y adaptadores) que separa el dominio del núcleo de los detalles de infraestructura, permitiendo mantener la testabilidad y la evolución independiente de cada capa.

== 2.2 Funcionalidad del producto

La plataforma opera sobre tres flujos comerciales fundamentales:

=== Flujo 1: Venta Directa (Comprador Minorista ↔ Agricultor)

#figure(
```text
Minorista ──> Navega catálogo ──> Encuentra producto ──> Chat directo
                                                                 │
                                                                 v
                                                    Acuerdan entrega/precio
                                                                 │
                                                                 v
                                                    Minorista califica (opcional)
```
)

El comprador minorista explora el catálogo del agricultor (productos, precios, ubicación) y puede contactarlo directamente a través de un chat abierto. No requiere match ni intermediación de la plataforma. El agricultor y el comprador acuerdan los detalles de la transacción fuera del sistema.

*A diferencia de plataformas de propósito general como Facebook Marketplace, las publicaciones en Milpa tienen caducidad:* cada producto publicado tiene una fecha de expiración basada en la temporalidad del cultivo, evitando publicaciones obsoletas y reflejando la naturaleza perecedera de los productos agrícolas.

=== Flujo 2: Solicitud de Abastecimiento (Comprador Mayorista → Agricultor)

#figure(
```text
Mayorista ──> Publica solicitud ──> Agricultores ofertan ──> Mayorista evalúa
                                      │                              │
                                      └── (geolocalización,         │
                                           precio, reputación)       │
                                                                     v
                                                              Match (Tinder)
                                                                     │
                                                                     v
                                                              Chat post-match
                                                                     │
                                                                  entrega
                                                                     │
                                                                     v
                                                          Calificación mutua
```
)

El comprador mayorista (detallista o corporativo) publica una solicitud de abastecimiento especificando producto, cantidad, unidad de medida, ubicación y fecha límite opcional. Los agricultores registrados pueden enviar ofertas indicando cantidad disponible, precio, tiempo de entrega y comentarios. El comprador evalúa las ofertas utilizando el sistema tipo Tinder, que prioriza cercanía geográfica, precio, tiempo de entrega y reputación del agricultor. Al aceptar una oferta se genera un Match, las demás ofertas se rechazan automáticamente, y se habilita el chat entre ambas partes para coordinar la entrega. Al finalizar, ambas partes se califican mutuamente.

=== Flujo 3: Oferta de Liquidación (Agricultor → Compradores)

#figure(
```text
Agricultor ──> Publica lote ──> ¿Restringir a mayoristas?
                                     │              │
                                     Sí             No
                                     │              │
                                     v              v
                              Solo mayoristas   Todos los compradores
                                     │              │
                                     └──────┬───────┘
                                            v
                                 Compradores interesados
                                            │
                                            v
                                 Agricultor asigna lote
```
)

El agricultor con excedente de cosecha puede publicar una oferta de liquidación ofreciendo un lote completo a un precio preferencial para evitar la pérdida total del producto. El agricultor decide si la oferta es visible para todos los compradores o exclusivamente para compradores mayoristas. El agricultor puede elegir entre asignar el lote al primer comprador interesado (FCFS) o seleccionar entre los interesados según su criterio.

=== Funcionalidades transversales

- Sistema de calificaciones mutuas (1 a 5 estrellas) después de cada transacción completada.
- Geolocalización de agricultores para búsqueda por cercanía.
- Motor de filtros por producto, ubicación, precio y disponibilidad.
- *Publicaciones con fecha de expiración* — a diferencia de marketplaces genéricos, los productos agrícolas tienen caducidad. Cada publicación incluye una fecha de expiración configurada por el agricultor, reflejando la disponibilidad real del cultivo.
- Moderación de productos y contenido por parte del administrador.
- Sistema de reportes de publicaciones fraudulentas.

== 2.3 Características de los usuarios

=== Agricultor

- Puede ser persona natural (agricultor individual) o persona jurídica (empresa agrícola).
- Necesita registrar su ubicación (departamento, municipio, coordenadas).
- Gestiona su propio catálogo de productos con imágenes, variedades y precios.
- Puede publicar ofertas de liquidación de excedentes.
- Responde a solicitudes de abastecimiento de compradores mayoristas.
- Conocimiento tecnológico básico a medio; uso frecuente de dispositivos móviles.

=== Comprador Minorista

- Consumidor final que busca productos frescos para consumo personal o familiar.
- Compra en pequeñas cantidades.
- Navega el catálogo y contacta directamente al agricultor vía chat.
- Conocimiento tecnológico básico.
- Uso frecuente de dispositivos móviles.

=== Comprador Mayorista Detallista

- Dueño de puesto de mercado, distribuidor local o pequeño intermediario.
- Compra en volumen para reventa al detal.
- Publica solicitudes de abastecimiento y también puede participar en ofertas de liquidación.
- Mayor volumen de transacciones que el minorista.

=== Comprador Mayorista Corporativo

- Supermercados, cadenas de distribución, procesadores de alimentos o restaurantes.
- Compra en altos volúmenes de forma recurrente.
- Publica solicitudes de abastecimiento con requerimientos específicos.
- Busca proveedores confiables y con capacidad de respuesta.

=== Administrador

- Gestiona y modera la plataforma.
- Aprueba o rechaza productos reportados.
- Suspende usuarios que incumplen las normas.
- Gestiona el catálogo de categorías y unidades de medida permitidas.

== 2.4 Restricciones

- Desarrollo en periodo limitado de Hackathon (aproximadamente 3 meses).
- Uso de tecnologías de código abierto.
- Compatibilidad con Android, iOS y navegadores modernos.
- Conectividad a Internet requerida (con consideraciones para zonas rurales con conectividad limitada).


== 2.5 Suposiciones y dependencias

- Disponibilidad de un servidor de despliegue con Docker.
- Disponibilidad de servicios de mapas (OpenStreetMap / Leaflet).
- Disponibilidad de conexión a Internet para usuarios.
- Los agricultores cuentan con un teléfono inteligente con cámara para fotografiar sus productos.
- Las direcciones de geolocalización se basan en la división política de Nicaragua (departamentos y municipios).

== 2.6 Evolución previsible del sistema

Las siguientes funcionalidades se consideran para iteraciones posteriores al MVP y quedan documentadas para conocimiento de la arquitectura:

#table(
  columns: 3,
  table.header("Funcionalidad", "Descripción", "Prioridad futura"),
  ["Verificación SMS + Badge"], ["Verificación de número telefónico vía SMS que otorga una insignia de confianza al agricultor."], ["Alta"],
  ["Cuentas multi-usuario"], ["Posibilidad de que una empresa agrícola agregue múltiples empleados con permisos diferenciados para gestionar la cuenta."], ["Media"],
  ["Notificaciones push/email"], ["Sistema de notificaciones para eventos clave: nuevas ofertas, match, mensajes, recordatorios de entrega."], ["Alta"],
  ["Favoritos"], ["Los compradores pueden guardar agricultores como favoritos para acceso rápido."], ["Baja"],
  ["Bloqueo automático por reportes"], ["Suspensión automática de cuentas que acumulan un número determinado de reportes confirmados, pendiente revisión manual."], ["Media"],
  ["Detección de duplicados"], ["Prevención automática de publicaciones duplicadas del mismo agricultor para el mismo producto en período corto."], ["Baja"],
  ["Publicaciones incompletas"], ["Detección y bloqueo automático de publicaciones con información insuficiente antes de su publicación."], ["Baja"],
  ["Cálculo automático de total"], ["Cálculo del costo total según cantidad solicitada y precio ofertado."], ["Baja"],
  ["Expiración automática"], ["Cierre automático de solicitudes de abastecimiento sin match después de un período definido (ej. 7 días)."], ["Media"],
  ["Panel de estadísticas"], ["Dashboard administrativo con reportes de uso, transacciones y tendencias del mercado."], ["Media"],
  ["Algoritmo de recomendación"], ["Motor de recomendación avanzado basado en historial de compras, calificaciones y cercanía."], ["Baja"],
)

#pagebreak()

= 3. Requisitos Específicos

== 3.1 Requisitos comunes de interfaces

=== 3.1.1 Interfaces de usuario

- Interfaz web responsiva (escritorio y tableta).
- Aplicación móvil nativa para Android e iOS (desarrollada en Flutter).
- Diseño adaptado a dispositivos móviles (mobile-first).
- Interfaz de chat en tiempo real.
- Visualización de mapas para geolocalización de agricultores.
- Interfaz tipo Tinder para evaluación de ofertas (mayorista).
- Navegación con búsqueda y filtros accesible para usuarios con conocimiento tecnológico básico.

=== 3.1.2 Interfaces de software

- API REST (Go) para operaciones CRUD y lógica de negocio.
- WebSocket para mensajería en tiempo real (chat).
- PostgreSQL como motor de base de datos relacional.
- Redis como caché de consultas frecuentes.
- Servicio de mapas OpenStreetMap con Leaflet.

=== 3.1.3 Interfaces de comunicación

- HTTPS para toda la comunicación con la API REST.
- JSON como formato de intercambio de datos.
- WebSocket sobre WSS para el chat.
- JWT para autenticación y autorización.

#pagebreak()

== 3.2 Requisitos Funcionales

=== RF-01: Registro de usuarios

#table(
  columns: (35%, 65%),
  [Número de requisito], [RF-01],
  [Nombre de requisito], [Registro de usuarios],
  [Tipo], [Funcional],
  [Prioridad], [Alta / Esencial],
  [MVP], [Sí],
)

El sistema deberá permitir el registro de usuarios mediante correo electrónico o número de teléfono válido. Cada usuario podrá tener únicamente una cuenta activa.

*Reglas de negocio asociadas:*
- El correo electrónico o número de teléfono debe ser único en el sistema.
- La contraseña se almacenará cifrada (bcrypt).
- El usuario debe aceptar los términos y condiciones antes de completar el registro.
- El usuario seleccionará su tipo de cuenta al registrarse: *Agricultor*, *Comprador Minorista*, *Comprador Mayorista Detallista*, *Comprador Mayorista Corporativo*.
- Los tipos de cuenta no son intercambiables (un agricultor no puede actuar como comprador en la misma cuenta, y viceversa).

=== RF-02: Autenticación y sesión

#table(
  columns: (35%, 65%),
  [Número de requisito], [RF-02],
  [Nombre de requisito], [Autenticación y sesión],
  [Tipo], [Funcional],
  [Prioridad], [Alta / Esencial],
  [MVP], [Sí],
)

El sistema deberá permitir la autenticación de usuarios registrados mediante credenciales (email/contraseña o teléfono/contraseña) y emitir tokens JWT para la gestión de sesión.

*Reglas de negocio asociadas:*
- Las contraseñas se verifican contra el hash almacenado (bcrypt).
- El token JWT incluirá el identificador del usuario, su rol y una fecha de expiración.
- Los tokens expirados serán rechazados por el sistema.

=== RF-03: Perfil de agricultor

#table(
  columns: (35%, 65%),
  [Número de requisito], [RF-03],
  [Nombre de requisito], [Perfil de agricultor],
  [Tipo], [Funcional],
  [Prioridad], [Alta / Esencial],
  [MVP], [Sí],
)

El sistema deberá permitir que un usuario registrado como agricultor complete su perfil con información obligatoria: nombre o nombre comercial, ubicación (departamento, municipio, dirección, coordenadas geográficas) y datos de contacto (teléfono, correo electrónico).

*Reglas de negocio asociadas:*
- Un agricultor no podrá publicar productos hasta completar su perfil.
- La ubicación se captura mediante selección de departamento/municipio y coordenadas geográficas (latitud, longitud).
- El agricultor puede optar por registrarse como persona natural o como empresa agrícola (ver RF-04).

=== RF-04: Registro y verificación de empresa agrícola

#table(
  columns: (35%, 65%),
  [Número de requisito], [RF-04],
  [Nombre de requisito], [Registro y verificación de empresa agrícola],
  [Tipo], [Funcional],
  [Prioridad], [Alta / Esencial],
  [MVP], [Sí],
)

El sistema deberá permitir que los agricultores que operan como empresa registren su información comercial completa. Las empresas verificadas recibirán una insignia de *Proveedor Verificado* que les otorgará prioridad en resultados de búsqueda.

*Reglas de negocio asociadas:*
- Solo las empresas registradas podrán publicar productos como proveedoras formales.
- Cada empresa debe registrarse con correo electrónico y teléfono únicos.
- El nombre comercial no podrá repetirse.
- Información obligatoria: nombre comercial, nombre del representante, número de identificación fiscal (RUC/NIT), dirección, municipio, departamento, teléfono, correo electrónico.
- El correo electrónico deberá verificarse antes de activar la cuenta.
- La empresa puede subir documentos para verificación (licencia, RUC, cédula del representante).
- Las empresas no verificadas pueden registrarse pero tendrán menor visibilidad o límite de publicaciones.
- El sistema asignará un identificador único a cada empresa registrada.
- El sistema almacenará el historial de cambios de la información de la empresa para auditoría.

*Diferencia MVP vs futuro:* La subida de documentos para verificación y el historial de auditoría se implementan desde el MVP. La verificación SMS con badge (RF-03, opcional para agricultores individuales) se deja como mejora futura.

=== RF-05: Gestión de catálogo de productos

#table(
  columns: (35%, 65%),
  [Número de requisito], [RF-05],
  [Nombre de requisito], [Gestión de catálogo de productos],
  [Tipo], [Funcional],
  [Prioridad], [Alta / Esencial],
  [MVP], [Sí],
)

El sistema deberá permitir que el agricultor cree, edite, active y desactive productos en su catálogo. El catálogo funciona como una vitrina accesible para todos los compradores, similar a un perfil de negocio en redes sociales, pero estructurado con datos comerciales.

*Reglas de negocio asociadas:*
- Cada producto incluirá: nombre del cultivo, variedad, descripción, precio por unidad de medida, unidad de medida (seleccionada de un catálogo predefinido), cantidad disponible, imágenes, ubicación del producto (puede diferir de la ubicación del agricultor).
- Las unidades de medida deben seguir un catálogo predefinido por la plataforma (quintales, libras, toneladas, unidades, docenas, etc.).
- El sistema evitará publicaciones duplicadas del mismo agricultor para el mismo producto en un período corto (definido por configuración).
- Cada publicación de producto tendrá una fecha de expiración configurada por el agricultor basada en la disponibilidad real del cultivo. Al expirar, la publicación pasará a estado inactivo y no será visible en el marketplace. El agricultor podrá renovarla si el producto sigue disponible.
- Solo podrán publicarse productos permitidos por la plataforma (ver RF-06).
- Si el producto contiene información incompleta, el sistema lo notificará al agricultor y evitará su publicación hasta que se complete.
- El agricultor puede indicar si realiza entregas a domicilio o si el producto debe retirarse en un punto específico.

=== RF-06: Catálogo de tipos de producto (administrador)

#table(
  columns: (35%, 65%),
  [Número de requisito], [RF-06],
  [Nombre de requisito], [Catálogo de tipos de producto],
  [Tipo], [Funcional],
  [Prioridad], [Alta / Esencial],
  [MVP], [Sí],
)

El sistema deberá proporcionar un catálogo administrable de tipos de producto permitidos en la plataforma. Solo los productos pertenecientes a categorías activas en este catálogo podrán ser publicados por los agricultores.

*Reglas de negocio asociadas:*
- El administrador puede crear, editar y desactivar tipos de producto.
- Cada tipo de producto incluye: nombre, descripción, categoría principal (frutales, cítricos, otros), unidad de medida predeterminada.
- El catálogo inicial se enfocará en productos frutales y cítricos, alineado con el reto del Hackathon.

=== RF-07: Marketplace y búsqueda

#table(
  columns: (35%, 65%),
  [Número de requisito], [RF-07],
  [Nombre de requisito], [Marketplace y búsqueda],
  [Tipo], [Funcional],
  [Prioridad], [Alta / Esencial],
  [MVP], [Sí],
)

El sistema deberá permitir a todos los compradores (minoristas y mayoristas) navegar el catálogo de productos publicado por los agricultores, con capacidades de búsqueda y filtrado.

*Reglas de negocio asociadas:*
- La búsqueda permitirá filtrar por: tipo de producto, ubicación (departamento/municipio), precio, agricultor.
- Los resultados se mostrarán con información básica: producto, precio, agricultor, ubicación, imagen.
- Los agricultores verificados tendrán prioridad en los resultados de búsqueda.
- La geolocalización se utilizará para ordenar resultados por cercanía (cuando el comprador proporcione su ubicación).

=== RF-08: Chat directo (minorista ↔ agricultor)

#table(
  columns: (35%, 65%),
  [Número de requisito], [RF-08],
  [Nombre de requisito], [Chat directo minorista ↔ agricultor],
  [Tipo], [Funcional],
  [Prioridad], [Alta / Esencial],
  [MVP], [Sí],
)

El sistema deberá proporcionar un canal de chat en tiempo real (WebSocket) entre compradores minoristas y agricultores, sin necesidad de match previo.

*Reglas de negocio asociadas:*
- Cualquier comprador minorista puede iniciar un chat con cualquier agricultor desde la página del producto.
- El chat es bidireccional y en tiempo real.
- El sistema almacenará el historial de conversaciones.
- No se requiere match ni aceptación para que el chat funcione en este flujo.

=== RF-09: Geolocalización

#table(
  columns: (35%, 65%),
  [Número de requisito], [RF-09],
  [Nombre de requisito], [Geolocalización],
  [Tipo], [Funcional],
  [Prioridad], [Media / Deseado],
  [MVP], [Sí — básico],
)

El sistema deberá capturar y almacenar coordenadas geográficas (latitud, longitud) de los agricultores y sus productos, y utilizarlas para visualización en mapa y ordenamiento por cercanía.

*Detalle:*
- Las coordenadas se capturan durante el registro del perfil del agricultor.
- El marketplace muestra la ubicación del agricultor en un mapa interactivo (OpenStreetMap + Leaflet).
- El sistema de match prioriza ofertas por cercanía geográfica entre el comprador mayorista y el agricultor.

=== RF-10: Publicación de solicitud de abastecimiento

#table(
  columns: (35%, 65%),
  [Número de requisito], [RF-10],
  [Nombre de requisito], [Publicación de solicitud de abastecimiento],
  [Tipo], [Funcional],
  [Prioridad], [Alta / Esencial],
  [MVP], [Sí],
)

El sistema deberá permitir que un comprador mayorista (detallista o corporativo) publique una solicitud de abastecimiento especificando su necesidad de producto.

*Reglas de negocio asociadas:*
- La solicitud debe incluir: producto, cantidad requerida, unidad de medida, ubicación, fecha límite (opcional).
- La publicación permanecerá activa hasta que el comprador la cierre, expire automáticamente (futuro) o se genere un match.
- El comprador podrá cancelar la solicitud siempre que aún no haya aceptado una oferta.
- Una solicitud no podrá modificarse después de aceptar una oferta.

=== RF-11: Oferta a solicitud de abastecimiento

#table(
  columns: (35%, 65%),
  [Número de requisito], [RF-11],
  [Nombre de requisito], [Oferta a solicitud de abastecimiento],
  [Tipo], [Funcional],
  [Prioridad], [Alta / Esencial],
  [MVP], [Sí],
)

El sistema deberá permitir que los agricultores registrados envíen ofertas en respuesta a una solicitud de abastecimiento publicada por un comprador mayorista.

*Reglas de negocio asociadas:*
- Un agricultor solo podrá ofertar si tiene disponible el producto solicitado (basado en su catálogo).
- Cada agricultor podrá enviar una única oferta por solicitud.
- La oferta debe incluir: cantidad disponible, precio, tiempo estimado de entrega, comentarios opcionales.
- El agricultor podrá editar su oferta mientras no haya sido aceptada.
- El agricultor podrá retirar su oferta antes de que sea aceptada.

=== RF-12: Match y ciclo de transacción

#table(
  columns: (35%, 65%),
  [Número de requisito], [RF-12],
  [Nombre de requisito], [Match y ciclo de transacción],
  [Tipo], [Funcional],
  [Prioridad], [Alta / Esencial],
  [MVP], [Sí],
)

El sistema deberá proporcionar un mecanismo tipo Tinder para que el comprador mayorista evalúe y acepte ofertas de agricultores, y gestione el ciclo de vida de la transacción resultante.

*Reglas de negocio asociadas:*
- Las ofertas se muestran priorizando cercanía geográfica, precio, tiempo de entrega y reputación del agricultor.
- El comprador puede deslizar o seleccionar ofertas para Aceptar o Descartar.
- Al aceptar una oferta se genera un *Match*.
- Una vez creado el Match, las demás ofertas pasan automáticamente al estado de rechazadas.
- Solo puede existir un Match activo por solicitud de abastecimiento.

*Ciclo de vida del Match/Transacción:*

#table(
  columns: 2,
  table.header("Estado", "Descripción"),
  [`matched`], ["Match creado. Comprador aceptó oferta. Chat habilitado. Coordinación en curso."],
  [`in_progress`], ["Ambas partes confirman que la transacción está activa (entrega en proceso)."],
  [`completed`], ["Ambas partes confirman que la entrega se realizó satisfactoriamente."],
  [`cancelled`], ["Cualquiera de las partes cancela antes de completarse. Se requiere indicar motivo."],
)

- Si se cancela antes de la entrega, ambas partes deben indicar el motivo.
- El sistema actualizará automáticamente la disponibilidad del producto cuando se confirme un match.
- Una solicitud se cierra automáticamente cuando se completa la cantidad solicitada.

=== RF-13: Chat post-match (mayorista ↔ agricultor)

#table(
  columns: (35%, 65%),
  [Número de requisito], [RF-13],
  [Nombre de requisito], [Chat post-match mayorista ↔ agricultor],
  [Tipo], [Funcional],
  [Prioridad], [Alta / Esencial],
  [MVP], [Sí],
)

El sistema deberá habilitar un canal de chat en tiempo real (WebSocket) entre el comprador mayorista y el agricultor únicamente después de que se haya generado un Match.

*Reglas de negocio asociadas:*
- El chat se habilita automáticamente al crearse el Match.
- El chat es bidireccional y en tiempo real.
- El sistema almacenará el historial de conversaciones.
- El chat permanece accesible incluso después de completada o cancelada la transacción.

=== RF-14: Publicación de oferta de liquidación

#table(
  columns: (35%, 65%),
  [Número de requisito], [RF-14],
  [Nombre de requisito], [Publicación de oferta de liquidación],
  [Tipo], [Funcional],
  [Prioridad], [Alta / Esencial],
  [MVP], [Sí],
)

El sistema deberá permitir que el agricultor publique ofertas de liquidación de lotes completos a precio preferencial, para evitar la pérdida total de excedentes de cosecha.

*Reglas de negocio asociadas:*
- La oferta debe incluir: producto, cantidad total del lote, precio total o por unidad, unidad de medida, tiempo estimado de entrega, ubicación.
- El agricultor puede restringir la visibilidad de la oferta a: todos los compradores, solo mayoristas detallistas, solo mayoristas corporativos, o mayoristas (ambos tipos).
- El agricultor puede elegir el método de asignación: primero en llegar (FCFS) o selección manual entre interesados.
- Si el agricultor selecciona asignación manual, puede ver la lista de compradores interesados con información básica y elegir.
- Una vez asignado el lote, la oferta se cierra.

=== RF-15: Calificaciones

#table(
  columns: (35%, 65%),
  [Número de requisito], [RF-15],
  [Nombre de requisito], [Calificaciones],
  [Tipo], [Funcional],
  [Prioridad], [Media / Deseado],
  [MVP], [Sí],
)

El sistema deberá permitir que compradores y agricultores se califiquen mutuamente después de finalizar una transacción (match completado en flujo mayorista, o transacción acordada en flujo minorista).

*Reglas de negocio asociadas:*
- La calificación es de 1 a 5 estrellas.
- Un usuario no puede calificarse a sí mismo.
- Solo se permite una calificación por transacción por usuario.
- La calificación es mutua: ambas partes califican a la contraparte.
- Las calificaciones son visibles en el perfil público del agricultor o comprador.
- El sistema muestra el promedio de calificaciones y el número total de calificaciones.

=== RF-16: Reportes y moderación

#table(
  columns: (35%, 65%),
  [Número de requisito], [RF-16],
  [Nombre de requisito], [Reportes y moderación],
  [Tipo], [Funcional],
  [Prioridad], [Media / Deseado],
  [MVP], [Sí — básico],
)

El sistema deberá permitir que los usuarios reporten publicaciones (productos, ofertas, solicitudes) o perfiles que consideren fraudulentos, falsos o que incumplan las normas de la plataforma.

*Reglas de negocio asociadas:*
- El administrador recibe los reportes y puede tomar acción: aprobar, rechazar o suspender la publicación/usuario.
- Los usuarios que acumulen múltiples reportes confirmados podrán ser suspendidos.
- El sistema registrará todas las operaciones de moderación para auditoría.

=== RF-17: Administración del sistema

#table(
  columns: (35%, 65%),
  [Número de requisito], [RF-17],
  [Nombre de requisito], [Administración del sistema],
  [Tipo], [Funcional],
  [Prioridad], [Alta / Esencial],
  [MVP], [Sí],
)

El sistema deberá proporcionar un panel de administración para gestionar usuarios, tipos de producto, unidades de medida, reportes y moderación de contenido.

*Funcionalidades del panel:*
- Gestión de usuarios (ver, suspender, activar).
- Gestión del catálogo de tipos de producto y unidades de medida.
- Gestión de reportes (bandeja de reportes con acciones de moderación).
- Visualización básica de actividad de la plataforma.

#pagebreak()

== 3.3 Requisitos No Funcionales

=== RNF-01: Rendimiento

#table(
  columns: (35%, 65%),
  [Número de requisito], [RNF-01],
  [Nombre de requisito], [Rendimiento],
  [Tipo], [Restricción],
  [Prioridad], [Alta / Esencial],
)

El sistema deberá responder a solicitudes de la API REST en un tiempo inferior a 2 segundos bajo condiciones normales de operación. Las consultas frecuentes (catálogo, búsqueda) deberán utilizar Redis como capa de caché para reducir la carga sobre PostgreSQL.

=== RNF-02: Seguridad

#table(
  columns: (35%, 65%),
  [Número de requisito], [RNF-02],
  [Nombre de requisito], [Seguridad],
  [Tipo], [Restricción],
  [Prioridad], [Alta / Esencial],
)

- Las contraseñas se almacenarán cifradas mediante bcrypt.
- Toda la comunicación será sobre HTTPS.
- La autenticación utilizará JWT con expiración.
- La información personal solo se compartirá entre comprador y agricultor después de un Match (flujo mayorista) o al iniciar un chat (flujo minorista).
- El sistema registrará todas las operaciones para fines de auditoría.
- Se implementará protección contra ataques comunes (SQL injection, XSS, CSRF).

=== RNF-03: Disponibilidad

#table(
  columns: (35%, 65%),
  [Número de requisito], [RNF-03],
  [Nombre de requisito], [Disponibilidad],
  [Tipo], [Restricción],
  [Prioridad], [Media / Deseado],
)

El sistema deberá mantener una disponibilidad mínima del 99% durante el período de evaluación del Hackathon.

=== RNF-04: Mantenibilidad

#table(
  columns: (35%, 65%),
  [Número de requisito], [RNF-04],
  [Nombre de requisito], [Mantenibilidad],
  [Tipo], [Restricción],
  [Prioridad], [Alta / Esencial],
)

El sistema se desarrollará bajo una arquitectura hexagonal (puertos y adaptadores) que separe el dominio del núcleo de los detalles de infraestructura, permitiendo la evolución independiente de cada capa y facilitando las pruebas unitarias.

=== RNF-05: Escalabilidad

#table(
  columns: (35%, 65%),
  [Número de requisito], [RNF-05],
  [Nombre de requisito], [Escalabilidad],
  [Tipo], [Restricción],
  [Prioridad], [Media / Deseado],
)

La arquitectura deberá permitir la incorporación de nuevos módulos y el incremento del número de usuarios sin requerir una reestructuración completa del sistema. La capa de caché (Redis) deberá soportar el crecimiento de consultas.

=== RNF-06: Compatibilidad

#table(
  columns: (35%, 65%),
  [Número de requisito], [RNF-06],
  [Nombre de requisito], [Compatibilidad],
  [Tipo], [Restricción],
  [Prioridad], [Alta / Esencial],
)

La API deberá ser consumible desde aplicaciones web (React) y móviles (Flutter Android e iOS). La comunicación será mediante JSON sobre HTTP/HTTPS y WebSocket.

=== RNF-07: Privacidad y protección de datos

#table(
  columns: (35%, 65%),
  [Número de requisito], [RNF-07],
  [Nombre de requisito], [Privacidad y protección de datos],
  [Tipo], [Restricción],
  [Prioridad], [Alta / Esencial],
)

- Toda la información personal será protegida y solo se compartirá entre comprador y agricultor cuando exista una interacción directa (chat o match).
- La plataforma no compartirá información de contacto con terceros no involucrados en una transacción.
- El sistema registrará todas las operaciones para fines de auditoría.

#pagebreak()

= 4. Apéndices

== A: Stack Tecnológico

#table(
  columns: (40%, 60%),
  table.header(
    [Componente],
    [Tecnología],
  ),
  [Lenguaje principal], [Go],
  [Framework API], [Chi (router)],
  [Base de datos], [PostgreSQL],
  [Caché], [Redis],
  [Autenticación], [JWT (golang-jwt)],
  [Chat], [WebSocket (gorilla/websocket)],
  [Mapas], [OpenStreetMap + Leaflet],
  [Contenedores], [Docker],
  [Control de versiones], [Git + GitHub],
  [Frontend Web], [React + TypeScript],
  [App Móvil], [Flutter],
  [Arquitectura], [Hexagonal (puertos y adaptadores)],
)

== B: Reglas de negocio detalladas

=== B.1 Usuarios

- Todo usuario debe registrarse con un correo electrónico o número de teléfono válido.
- Cada usuario podrá tener únicamente una cuenta activa.
- Los tipos de cuenta son mutuamente excluyentes: Agricultor, Comprador Minorista, Comprador Mayorista Detallista, Comprador Mayorista Corporativo.
- Los agricultores deberán completar su perfil con nombre, ubicación y datos de contacto antes de publicar productos.
- Los usuarios podrán ser suspendidos por incumplir las normas de la plataforma.
- Los correos electrónicos y números de teléfono deben ser únicos.

=== B.2 Agricultores (individuales y empresas)

- Los agricultores pueden operar como persona natural o como empresa registrada.
- Las empresas deben proporcionar información fiscal (RUC/NIT) para verificación.
- Las empresas verificadas reciben insignia de "Proveedor Verificado" y prioridad en búsquedas.
- Las empresas no verificadas pueden registrarse pero con visibilidad reducida.
- El sistema asignará un identificador único a cada empresa registrada.
- Una empresa podrá pertenecer a una categoría principal (Agrícola, Ferretería, Alimentos, Construcción, etc.), aunque podrá ofrecer varios productos.
- Si una empresa acumula reportes confirmados, su cuenta pasará a revisión.
- Una empresa podrá eliminar su cuenta solo si no tiene transacciones pendientes.
- El sistema almacenará el historial de cambios de la información de la empresa.
- La plataforma verificará que el teléfono y correo electrónico no estén asociados a otra empresa activa.

=== B.3 Publicaciones de demanda (solicitudes de abastecimiento)

- Un comprador mayorista deberá especificar: producto, cantidad requerida, unidad de medida, ubicación y fecha límite (opcional).
- Una publicación permanecerá activa hasta que el comprador la cierre, expire o se genere un match.
- Un comprador podrá cancelar su solicitud siempre que aún no haya aceptado una oferta.
- Una publicación no podrá modificarse después de aceptar una oferta.

=== B.4 Ofertas de agricultores

- Un agricultor solo podrá ofertar si dispone del producto solicitado en su catálogo.
- Cada agricultor podrá enviar una única oferta por publicación (editable mientras no sea aceptada).
- Una oferta debe incluir: cantidad disponible, precio, tiempo estimado de entrega y comentarios opcionales.
- Un agricultor podrá retirar su oferta antes de que sea aceptada.

=== B.5 Match

- Se genera un Match cuando el comprador acepta una oferta.
- Una vez creado el Match, las demás ofertas pasan automáticamente al estado de rechazadas.
- Solo puede existir un Match activo por publicación.
- El chat se habilita automáticamente al generarse el Match.
- Si la transacción es cancelada antes de la entrega, ambas partes deben indicar el motivo.

=== B.6 Chat

- El chat directo (minorista ↔ agricultor) está disponible sin restricciones.
- El chat post-match (mayorista ↔ agricultor) se habilita solo después del Match.
- El sistema almacenará el historial de conversaciones.

=== B.7 Productos

- Solo podrán publicarse productos permitidos por la plataforma (catálogo administrable).
- Las unidades de medida deben seguir un catálogo predefinido.
- El sistema evitará publicaciones duplicadas del mismo agricultor para el mismo producto en período corto.
- Cada publicación de producto tiene una fecha de expiración. Al expirar, la publicación se oculta automáticamente del marketplace.
- El agricultor puede renovar una publicación expirada si el producto sigue disponible.
- Productos con información incompleta no podrán publicarse.
- El agricultor puede indicar si realiza entregas a domicilio o el producto se retira en punto específico.

=== B.8 Calificaciones

- Compradores y agricultores pueden calificarse después de finalizar la transacción.
- La calificación es de 1 a 5 estrellas.
- Un usuario no puede calificarse a sí mismo.
- Solo se permite una calificación por transacción por usuario.
- Las calificaciones son mutuas y visibles en el perfil público.

=== B.9 Ofertas de liquidación

- El agricultor puede restringir la visibilidad de la oferta por tipo de comprador.
- El agricultor elige el método de asignación: FCFS o selección manual.
- Una vez asignado, la oferta se cierra automáticamente.

=== B.10 Disponibilidad

- Si un agricultor indica una cantidad menor a la solicitada, el comprador puede aceptarla de forma parcial.
- El sistema actualizará automáticamente la disponibilidad cuando se confirme una venta.
- Una solicitud se cierra automáticamente cuando se completa la cantidad solicitada.

=== B.11 Algoritmo de match (Tinder)

- Las ofertas se muestran priorizando cercanía geográfica.
- El sistema ordena ofertas por precio, distancia, tiempo de entrega o reputación del agricultor.
- Los compradores pueden deslizar o seleccionar ofertas para Aceptar o Descartar.
- El sistema recomienda agricultores con mejor historial y calificaciones.

=== B.12 Seguridad y auditoría

- Toda la información personal se protege y solo se comparte entre partes con interacción directa.
- El sistema registra todas las operaciones para auditoría.
- Usuarios con múltiples reportes pueden ser bloqueados automáticamente hasta revisión (futuro).

=== B.13 Reportes

- Los usuarios pueden reportar publicaciones falsas o fraudulentas.
- El sistema detecta publicaciones con información incompleta (futuro: bloqueo automático).

== C: Modelo de dominio conceptual

El modelo de dominio refleja las entidades principales del sistema Milpa. Las entidades marcadas con asterisco (`*`) existen actualmente en la base de código y serán adaptadas. Las marcadas con (`+`) son nuevas para este reto.

=== Entidades actuales (a adaptar)

#table(
  columns: (25%, 50%, 25%),
  table.header("Entidad", "Descripción", "Ajuste para Milpa"),
  ["User (\*)"], ["Usuario del sistema con rol."], ["Los roles cambian de MIPYME/Provider a Agricultor/CompradorMinorista/CompradorMayoristaDetallista/CompradorMayoristaCorporativo/Admin."],
  ["Company (\*)"], ["Empresa registrada con información fiscal."], ["Mapea a empresa agrícola. Se añade soporte para agricultores individuales sin empresa."],
  ["Address (\*)"], ["Dirección con coordenadas geográficas."], ["Se mantiene. Se añade relación directa con agricultor (no solo con empresa)."],
  ["Offering (\*)"], ["Producto o servicio publicado."], ["Pasa a ser exclusivamente producto agrícola. Se añade variedad, unidad de medida, cantidad disponible."],
  ["Review (\*)"], ["Calificación y comentario."], ["Se mantiene. Se generaliza para que tanto agricultores como compradores puedan calificar y ser calificados."],
)

=== Nuevas entidades

#table(
  columns: (25%, 75%),
  table.header("Entidad", "Descripción"),
  ["SupplyRequest (+)"], ["Solicitud de abastecimiento publicada por un comprador mayorista: producto, cantidad, unidad, ubicación, fecha límite."],
  ["SupplyOffer (+)"], ["Oferta de un agricultor a una solicitud de abastecimiento: cantidad, precio, tiempo de entrega, comentarios."],
  ["Match (+)"], ["Vínculo generado cuando el comprador acepta una oferta. Contiene el estado de la transacción (matched, in_progress, completed, cancelled) y el motivo de cancelación."],
  ["ChatMessage (+)"], ["Mensaje individual dentro de una conversación de chat. Asociado a un Match (flujo mayorista) o a un chat directo (flujo minorista)."],
  ["Notification (+)"], ["Notificación de evento: nueva oferta, match, mensaje, recordatorio de entrega. (Futuro)"],
  ["LiquidationOffer (+)"], ["Oferta de liquidación de lote publicada por el agricultor: producto, cantidad, precio, restricción de visibilidad, método de asignación."],
  ["Report (+)"], ["Reporte de publicación o perfil realizado por un usuario."],
  ["Favorite (+)"], ["Relación de favorito entre un comprador y un agricultor. (Futuro)"],
)

#pagebreak()

== Historial del documento

#table(
  columns: 4,
  table.header("Fecha", "Versión", "Autor", "Cambios"),
  [5/Jul/26], [1.0], [Oscar Reyes], ["Documento inicial (basado en reto Plataforma de Proveedores)."],
  [26/Jul/26], [2.0], [Oscar Reyes], ["Reescritura completa: nuevo reto Eco-Mercado Digital, nombre Milpa, tres flujos comerciales, nuevo modelo de usuarios, reglas de negocio actualizadas."],
)
