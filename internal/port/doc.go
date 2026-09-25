/*
Package port holds the contracts the application needs from outside
services that are not data storage: token signing, password hashing,
categorization, and receipt vision. Storage contracts live in the
repository package instead. Implementations live in infrastructure,
and the usecase layer depends only on these interfaces.
*/
package port
