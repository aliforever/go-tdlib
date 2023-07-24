#include <td/telegram/td_log.h>
#include <td/telegram/td_json_client.h>

extern void go_td_log_message_callback_ptr(int verbosity_level, char *message);

void go_td_log_message_callback_ptr_proxy(int verbosity_level, const char *message);
void set_log_message_callback(int max_verbosity_level);