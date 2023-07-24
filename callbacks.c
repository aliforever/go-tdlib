#include "callbacks.h"

extern void go_td_log_message_callback_ptr(int verbosity_level, char *message);

void go_td_log_message_callback_ptr_proxy(int verbosity_level, const char *message) {
    go_td_log_message_callback_ptr(verbosity_level, (char *)message);
}

void set_log_message_callback(int max_verbosity_level) {
    td_set_log_message_callback(max_verbosity_level, go_td_log_message_callback_ptr_proxy);
}