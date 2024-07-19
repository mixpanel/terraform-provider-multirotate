resource "multirotate_set" "token_rotation" {
  rotation_period = "24h"
  number          = "2"
}