# Skipping Lunch

![lunch](https://static0.gamerantimages.com/wordpress/wp-content/uploads/2021/07/I-Think-You-Should-Leave-hot-dog-1.jpg)

We pushed lunch back to 1:30, so that Dennis could make the earlier flight.
We're meeting now to discuss the dreaded reorg.

[prompt]:# (cant_skip "Can you skip lunch?" [y n] n)

```verify
if [[ "$cant_skip" == "y" ]]; then
  exit 1
else
  exit 0
fi
```

You can't skip lunch.
