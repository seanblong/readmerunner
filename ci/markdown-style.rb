# frozen_string_literal: true

################################################################################
# Style file for markdownlint.
#
# https://github.com/markdownlint/markdownlint/blob/master/docs/configuration.md
#
# This file is referenced by the project `.mdlrc`.
################################################################################

#===============================================================================
# Start with all built-in rules.
# https://github.com/markdownlint/markdownlint/blob/master/docs/RULES.md
all

#===============================================================================
# Override default parameters for some built-in rules.
# https://github.com/markdownlint/markdownlint/blob/master/docs/creating_styles.md#parameters


# Turn off code block rule.
exclude_rule 'MD046'

# Replace standard line-length rule with custom one.  Ignore tables, code blocks,
# and lines starting with "[embed]".
exclude_rule 'MD013'
rule 'custom-line-length', ignore_code_blocks: true, tables: false, ignore_prefix: '[embedmd]'

# Replace fenced code blocks should be surrounded by blank lines with custom one
# where the prefix "[embedmd]" is ignored from the rule.
exclude_rule 'MD031'
rule 'custom-blanks-around-fences', ignore_prefix: '[embedmd]'

# Allow tabs in code blocks
rule 'MD010', ignore_code_blocks: true

# Allow duplicate header names when nested
rule 'MD024', allow_different_nesting: true


# IMHO it's easier to read lists like:
# * outmost indent
#   - one indent
#   - second indent
# * Another major bullet
exclude_rule 'MD004' # Unordered list style

# I prefer two blank lines before each heading.
exclude_rule 'MD012' # Multiple consecutive blank lines

# I find it necessary to use '<br/>' to force line breaks.
exclude_rule 'MD033' # Inline HTML

# If a page is printed, it helps if the URL is viewable.
exclude_rule 'MD034' # Bare URL used

#===============================================================================
# Exclude rules for pragmatic reasons.

# Either disable this one or MD024 - Multiple headers with the same content.
exclude_rule 'MD036' # Emphasis used instead of a header
