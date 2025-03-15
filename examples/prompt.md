# Egg Game

![](https://res.cloudinary.com/teepublic/image/private/s--7nI0aM0b--/c_crop,x_10,y_10/c_fit,h_945/c_crop,g_north_west,h_1260,w_945,x_-126,y_-158/co_rgb:cccaca,e_colorize,u_Misc:One%20Pixel%20Gray/c_scale,g_north_west,h_1260,w_945/fl_layer_apply,g_north_west,x_-126,y_-158/bo_105px_solid_white/e_overlay,fl_layer_apply,h_1260,l_Misc:Art%20Print%20Bumpmap,w_945/e_shadow,x_6,y_6/c_limit,h_1254,w_1254/c_lpad,g_center,h_1260,w_1260/b_rgb:eeeeee/c_limit,f_auto,h_630,q_auto:good:420,w_630/v1690474290/production/designs/48451700_1.jpg)


"Dude, you ran out of eggs."

[prompt]:# (eggprompt "Want to buy more eggs?" [yes] yes)

[prompt]:# (eggs "How many eggs?" [0 40 80] 80)

```bash
if [ "$eggs" = "80" ]; then
  echo "That's a lot of eggs big boy!"
else
  echo "You have $eggs eggs"
fi
```
